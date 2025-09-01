package pkg

import (
	"github.com/Bendomey/fincore-engine/internal/config"
	"github.com/Bendomey/fincore-engine/internal/handlers"
	"github.com/Bendomey/fincore-engine/internal/repository"
	"github.com/Bendomey/fincore-engine/internal/services"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type AppContext struct {
	DB         *gorm.DB
	Config     config.Config
	Repository repository.Repository
	Handlers   handlers.Handlers
	Services   services.Services
	Validator  *validator.Validate
}
