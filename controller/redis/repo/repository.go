package repo

import "github.com/uzzeet/uzzeet-gateway/models"

type CompositeRepository interface {
	GetRedisByID(models.CompositeID) (models.Composite, error)
}
