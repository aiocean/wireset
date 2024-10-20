package plan

import (
	"github.com/aiocean/wireset/feature/shopifyapp/models"
)

type Registry struct {
	PlanRepo *PlanRepository
}

// AddPlans registers a plan with the registry, if the plan already exists, it will not be registered again.
func (r *Registry) AddPlans(plans ...*models.Plan) error {
	for _, plan := range plans {
		if exists, err := r.PlanRepo.IsPlanExists(plan.ID); err != nil {
			return err
		} else if exists {
			return nil
		}

		if err := r.PlanRepo.CreatePlan(plan); err != nil {
			return err
		}
	}

	return nil
}
