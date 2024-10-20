package plan

import (
	"context"
	"errors"

	"github.com/aiocean/wireset/feature/shopifyapp/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrPlanNotFound = errors.New("plan not found")
	ErrNoPlanFound  = errors.New("no plan found for shop")
)

const (
	DatabaseName   = "plan"
	CollectionName = "plans"
)

func NewPlanRepository(mongoClient *mongo.Client) *PlanRepository {
	db := mongoClient.Database(DatabaseName)
	
	return &PlanRepository{
		PlansCollection: db.Collection(CollectionName),
	}
}

// MemoryPlanRepository is an in-memory implementation of PlanRepository.
type PlanRepository struct {
	PlansCollection *mongo.Collection
}

// GetPlan returns the pricing plan with the given ID.
func (r *PlanRepository) GetPlanByID(ID string) (*models.Plan, error) {
	var plan models.Plan
	if err := r.PlansCollection.FindOne(context.Background(), bson.M{"id": ID}).Decode(&plan); err != nil {
		return nil, err
	}

	return &plan, nil
}

// IsPlanExists checks if the plan with the given ID exists.
func (r *PlanRepository) IsPlanExists(ID string) (bool, error) {
	var plan models.Plan
	if err := r.PlansCollection.FindOne(context.Background(), bson.M{"id": ID}).Decode(&plan); err != nil {

		if err == mongo.ErrNoDocuments {
			return false, nil
		}

		return false, err
	}
	return true, nil
}

// ListPlans returns a list of all pricing plans.
func (r *PlanRepository) ListPlans() ([]*models.Plan, error) {
	var plans []*models.Plan
	cursor, err := r.PlansCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var plan models.Plan
		if err := cursor.Decode(&plan); err != nil {
			return nil, err
		}
		plans = append(plans, &plan)
	}
	return plans, nil
}

// CreatePlan creates a new pricing plan.
func (r *PlanRepository) CreatePlan(plan *models.Plan) error {
	_, err := r.PlansCollection.InsertOne(context.Background(), plan)
	return err
}

// UpdatePlan updates an existing pricing plan.
func (r *PlanRepository) UpdatePlan(plan *models.Plan) error {
	_, err := r.PlansCollection.ReplaceOne(context.Background(), bson.M{"id": plan.ID}, plan)
	return err
}

// DeletePlan deletes a pricing plan.
func (r *PlanRepository) DeletePlan(ID string) error {
	_, err := r.PlansCollection.DeleteOne(context.Background(), bson.M{"id": ID})
	return err
}

// GetFeaturesForPlan returns a list of features that are included in the given pricing plan.
func (r *PlanRepository) GetFeaturesForPlan(ID string) ([]*models.Feature, error) {
	var plan models.Plan
	if err := r.PlansCollection.FindOne(context.Background(), bson.M{"id": ID}).Decode(&plan); err != nil {
		return nil, err
	}
	return plan.Features, nil
}

// CanPlanFeature checks if the given plan ID has the given feature ID.
func (r *PlanRepository) CanPlanFeature(planID, featureID string) (bool, error) {
	var plan models.Plan
	if err := r.PlansCollection.FindOne(context.Background(), bson.M{"id": planID}).Decode(&plan); err != nil {
		return false, err
	}
	for _, feature := range plan.Features {
		if feature.ID == featureID {
			return true, nil
		}
	}
	return false, nil
}

// GetPlansOfShop returns a list of pricing plans for the given shop ID.
func (r *PlanRepository) GetPlansOfShop(shopID string) ([]*models.Plan, error) {
	var plans []*models.Plan
	var plan models.Plan
	if err := r.PlansCollection.FindOne(context.Background(), bson.M{"id": shopID}).Decode(&plan); err != nil {
		return nil, err
	}
	plans = append(plans, &plan)
	if len(plans) == 0 {
		return nil, ErrNoPlanFound
	}
	return plans, nil
}

// CanShopFeature checks if the given shop ID has the given feature ID.
func (r *PlanRepository) CanShopFeature(shopID, featureID string) (bool, error) {
	var plan models.Plan
	if err := r.PlansCollection.FindOne(context.Background(), bson.M{"id": shopID}).Decode(&plan); err != nil {
		return false, err
	}
	if plan.ID == "" {
		return false, ErrNoPlanFound
	}
	return r.CanPlanFeature(plan.ID, featureID)
}
