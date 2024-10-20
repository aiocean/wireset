package plan

import (
	"github.com/aiocean/wireset/feature/shopifyapp/models"
	"github.com/alitto/pond"
	"go.uber.org/zap"
)

type Registry struct {
	PlanRepo *PlanRepository
	WorkerPool *pond.WorkerPool
	Logger     *zap.Logger
}

// AddPlans registers plans with the registry, if a plan already exists, it will not be registered again.
func (r *Registry) AddPlans(plans ...*models.Plan) error {

	r.WorkerPool.Submit(func() {
		for _, plan := range plans {
			plan := plan // Capture loop variable
			r.WorkerPool.Submit(func() {
				exists, err := r.PlanRepo.IsPlanExists(plan.ID)
				if err != nil || exists {
					return
				}

				if err := r.PlanRepo.CreatePlan(plan); err != nil {
					r.Logger.Error("failed to create plan", zap.Error(err))
				}
			})
		}
	})

	return nil
}
