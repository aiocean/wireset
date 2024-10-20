package models

import "github.com/aiocean/wireset/feature/realtime/models"

const TopicSetActivateSubscription models.WebsocketTopic = "setActiveSubscription"

const SubscriptionStatusActive = "ACTIVE"
const SubscriptionStatusCancelled = "CANCELLED"
const SubscriptionStatusDeclined = "DECLINED"
const SubscriptionStatusExpired = "EXPIRED"
const SubscriptionStatusFrozen = "FROZEN"
const SubscriptionStatusPending = "PENDING"

/*
Status enum:
ACTIVE: The app subscription has been approved by the merchant. Active app subscriptions are billed to the shop. After payment, partners receive payouts.

CANCELLED: The app subscription was cancelled by the app. This could be caused by the app being uninstalled, a new app subscription being activated, or a direct cancellation by the app. This is a terminal state.

DECLINED: The app subscription was declined by the merchant. This is a terminal state.

EXPIRED: The app subscription wasn't approved by the merchant within two days of being created. This is a terminal state.

FROZEN: The app subscription is on hold due to non-payment. The subscription re-activates after payments resume.

PENDING: The app subscription is pending approval by the merchant.

ACCEPTED: The app subscription has been approved by the merchant and is ready to be activated by the app. As of API version 2021-01, when a merchant approves an app subscription, the status immediately transitions from pending to active.
*/
type SetActivateSubscriptionPayload struct {
	TrialDays int        `json:"trialDays"`
	Status    string     `json:"status"` // ACTIVE, UNPAID, EXPIRED
	Plan      *Plan      `json:"plan"`
}

const TopicNavigateTo models.WebsocketTopic = "navigateTo"

type NavigateToPayload struct {
	URL string `json:"url"`
}