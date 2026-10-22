You are helping the user integrate with the **FinCore Engine API** — a double-entry accounting REST API.

When the user asks you to interact with FinCore, use the reference below to write correct code or curl commands. Always use the production base URL unless the user specifies otherwise. Remind the user to store their `client_secret` securely — it's only returned once at registration.

---

# FinCore Engine API Reference

## Base URLs

| Environment | URL |
|-------------|-----|
| Production  | https://fincore-engine.fly.dev |
| Staging     | https://fincore-engine.fly.dev |
| Local dev   | http://localhost:5002 |

## Authentication

All requests (except client registration) require:
- `X-FinCore-Client-Id: <client_id>`
- `X-FinCore-Client-Secret: <client_secret>`

Register to get credentials:
```sh
curl -X POST https://fincore-engine.fly.dev/api/v1/clients \
  -H "Content-Type: application/json" \
  -d '{"name": "My App", "email": "myapp@example.com"}'
# Returns client_id and client_secret — save the secret, it's shown only once
```

## Response Patterns

- Single resource: `{"data": {...}}`
- List: `{"data": [...], "meta": {"page", "page_size", "order", "order_by", "total", "has_next_page", "has_previous_page"}}`
- Error: `{"errors": {"message": "..."}}`

## List Query Parameters

`page`, `page_size`, `order` (asc/desc), `order_by`, `query`, `search_fields`, `start_date`, `end_date`, `populate`

`order_by`, `search_fields` and `populate` only accept the values listed per endpoint below; anything else returns 400.

---

## Clients API

### POST /api/v1/clients — Register (no auth)
```json
{ "name": "string (required)", "email": "string (required)" }
```

### GET /api/v1/clients/me — Get current client (auth required)

---

## Accounts API

Account types: `ASSET`, `LIABILITY`, `EQUITY`, `INCOME`, `EXPENSE`

### POST /api/v1/accounts
```json
{
  "name": "string (required, 3-255)",
  "type": "ASSET|LIABILITY|EQUITY|INCOME|EXPENSE (required)",
  "is_contra": "boolean (required)",
  "is_group": "boolean (required)",
  "parent_account_id": "uuid (optional)",
  "description": "string (optional, 3-255)"
}
```

Codes are generated per client: the first ASSET account is `1000`, then `1001`, and so on (LIABILITY 2000, EQUITY 3000, INCOME 4000, EXPENSE 5000).

### GET /api/v1/accounts
Extra filters: `account_type`, `is_contra`, `is_group`, `parent_account_id`
Populate: `ParentAccount`
`order_by`: `created_at`, `updated_at`, `code`, `name`, `type` · `search_fields`: `code`, `name`, `description`

### GET /api/v1/accounts/{account_id}
Populate: `ParentAccount`

### PATCH /api/v1/accounts/{account_id}
```json
{ "name": "string (optional)", "description": "string (optional)" }
```
Fields you omit are left unchanged.

### GET /api/v1/accounts/{account_id}/ledger
Posted lines on the account in transaction-date order, with `opening_balance`, `closing_balance` and a `running_balance` per line (on the account's normal side, continuing across pages).
Params: `start_date`, `end_date` (YYYY-MM-DD or RFC3339; a date-only `end_date` includes that day), `page`, `page_size` (default 50, max 200). 404 for group accounts.

### DELETE /api/v1/accounts/{account_id}
- Cannot delete a group account with child accounts
- Cannot delete an account with journal entry lines

---

## Journal Entries API

Lifecycle: `DRAFT` → `POSTED`
Rule: **sum of debits must equal sum of credits across all lines**

### POST /api/v1/journal-entries
```json
{
  "status": "DRAFT|POSTED (required)",
  "reference": "string (required, 3-255, unique per client)",
  "transaction_date": "YYYY-MM-DD or RFC3339 (optional, defaults to now)",
  "metadata": "object (optional)",
  "lines": [
    {
      "account_id": "uuid (required)",
      "debit": "number >= 0 (required)",
      "credit": "number >= 0 (required)",
      "notes": "string (optional, 3-255)"
    }
  ]
}
```
Rules: at least 2 lines; each line has exactly one of `debit` or `credit` above zero; the total can't be zero; accounts must be your own non-group accounts.

A reference you've already used returns **409** with the existing entry's id in `errors.journal_entry_id`, so a retried create can treat 409 as success.

### GET /api/v1/journal-entries
Extra filters: `status`, `account_id` (entries with a line on that account), `reference` (exact)
Populate: `JournalEntryLines`, `Account` (lines with their accounts)
`order_by`: `created_at`, `updated_at`, `transaction_date`, `posted_at`, `reference` · `search_fields`: `reference`

### GET /api/v1/journal-entries/{journal_entry_id}

### PATCH /api/v1/journal-entries/{journal_entry_id}
Only DRAFT entries can be updated.
```json
{
  "reference": "string (optional)",
  "transaction_date": "YYYY-MM-DD (optional)",
  "metadata": "object (optional)",
  "lines": [
    {
      "id": "uuid (optional — include to update existing line)",
      "account_id": "uuid (optional)",
      "debit": "number (optional)",
      "credit": "number (optional)",
      "notes": "string (optional)"
    }
  ]
}
```

### PATCH /api/v1/journal-entries/{journal_entry_id}/post
Finalizes the entry. No request body. Only DRAFT entries.

### DELETE /api/v1/journal-entries/{journal_entry_id}
Only DRAFT entries can be deleted.

---

## Reports API

### GET /api/v1/reports/trial-balance
Every non-group account with `total_debit`, `total_credit`, `normal_balance` and `balance`, plus grand totals. Only POSTED entries count. ASSET and EXPENSE are debit-normal, others credit-normal; contra accounts are the opposite.
Param: `as_of` (YYYY-MM-DD includes that day, or RFC3339; defaults to now).

To reverse a posted entry, post a new entry with the debits and credits swapped.

---

## Example: Record a $500 Cash Sale

```sh
# 1. Create Cash account
curl -X POST https://fincore-engine.fly.dev/api/v1/accounts \
  -H "Content-Type: application/json" \
  -H "X-FinCore-Client-Id: $CLIENT_ID" \
  -H "X-FinCore-Client-Secret: $CLIENT_SECRET" \
  -d '{"name":"Cash","type":"ASSET","is_contra":false,"is_group":false}'

# 2. Create Sales Revenue account
curl -X POST https://fincore-engine.fly.dev/api/v1/accounts \
  -H "Content-Type: application/json" \
  -H "X-FinCore-Client-Id: $CLIENT_ID" \
  -H "X-FinCore-Client-Secret: $CLIENT_SECRET" \
  -d '{"name":"Sales Revenue","type":"INCOME","is_contra":false,"is_group":false}'

# 3. Create journal entry (50000 = $500.00 in cents)
curl -X POST https://fincore-engine.fly.dev/api/v1/journal-entries \
  -H "Content-Type: application/json" \
  -H "X-FinCore-Client-Id: $CLIENT_ID" \
  -H "X-FinCore-Client-Secret: $CLIENT_SECRET" \
  -d "{
    \"status\": \"DRAFT\",
    \"reference\": \"SALE-001\",
    \"transaction_date\": \"2024-01-15\",
    \"lines\": [
      {\"account_id\": \"$CASH_ID\", \"debit\": 50000, \"credit\": 0},
      {\"account_id\": \"$REVENUE_ID\", \"debit\": 0, \"credit\": 50000}
    ]
  }"

# 4. Post the entry
curl -X PATCH https://fincore-engine.fly.dev/api/v1/journal-entries/$ENTRY_ID/post \
  -H "Content-Type: application/json" \
  -H "X-FinCore-Client-Id: $CLIENT_ID" \
  -H "X-FinCore-Client-Secret: $CLIENT_SECRET"
```

---

## Common Errors

| Status | Meaning |
|--------|---------|
| 400 | Bad request — see `errors.message` |
| 401 | Missing or invalid auth headers |
| 404 | Not found, or belongs to another client |
| 409 | Journal entry reference already used — `errors.journal_entry_id` is the existing entry |
| 422 | Validation error |
| 429 | Rate limited: 50/s per client; 100/min per IP for requests that don't authenticate; failed auths 20/min per IP and client |
| 500 | Internal server error |

## Full Reference

- Complete reference: https://fincore-engine.fly.dev/llms-full.txt
- OpenAPI spec: https://fincore-engine.fly.dev/swagger/index.yaml
