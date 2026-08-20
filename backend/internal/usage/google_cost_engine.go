package usage

import (
	"sort"
	"time"
)

type GoogleCostEvent struct {
	ID, UserID, Operation string
	CreatedAt             time.Time
	Success               bool
	RequestCount          int
}

type GoogleCostRow struct {
	TotalRequests, SuccessfulRequests, FailedRequests, BillableRequests int
	FreeUsageCap, FreeUsageConsumed, PaidRequestCount                   int
	GrossCostMicros, EstimatedPayableMicros                             int64
	UserPayableMicros                                                   map[string]int64
}

// AllocateGoogleCost applies one verified free cap per Google SKU and
// calendar month. Event ID makes same-timestamp ordering stable.
func AllocateGoogleCost(events []GoogleCostEvent, pricing Pricing, from, to time.Time) map[string]GoogleCostRow {
	byOperation := map[string][]GoogleCostEvent{}
	for _, event := range events {
		byOperation[event.Operation] = append(byOperation[event.Operation], event)
	}
	result := map[string]GoogleCostRow{}
	for operation, operationEvents := range byOperation {
		sku, ok := pricing.GoogleSKUs[operation]
		if !ok || !sku.Verified {
			continue
		}
		sort.SliceStable(operationEvents, func(i, j int) bool {
			if operationEvents[i].CreatedAt.Equal(operationEvents[j].CreatedAt) {
				return operationEvents[i].ID < operationEvents[j].ID
			}
			return operationEvents[i].CreatedAt.Before(operationEvents[j].CreatedAt)
		})
		row := GoogleCostRow{FreeUsageCap: sku.FreeMonthly, UserPayableMicros: map[string]int64{}}
		eligible := 0
		for _, event := range operationEvents {
			inPeriod := (from.IsZero() || !event.CreatedAt.Before(from)) && (to.IsZero() || event.CreatedAt.Before(to))
			count := event.RequestCount
			if count <= 0 {
				count = 1
			}
			if inPeriod {
				row.TotalRequests += count
				if event.Success {
					row.SuccessfulRequests += count
				} else {
					row.FailedRequests += count
				}
			}
			if !event.Success {
				continue
			}
			for n := 0; n < count; n++ {
				eligible++
				if eligible <= sku.FreeMonthly {
					if inPeriod {
						row.FreeUsageConsumed++
					}
					continue
				}
				if inPeriod {
					row.BillableRequests++
					row.PaidRequestCount++
					row.UserPayableMicros[event.UserID] += sku.PriceMicrosPer1000 / 1000
					row.GrossCostMicros += sku.PriceMicrosPer1000 / 1000
					row.EstimatedPayableMicros += sku.PriceMicrosPer1000 / 1000
				}
			}
		}
		// Gross cost is based on successful selected-period events, including
		// those covered by the project free tier.
		selectedSuccessful := row.SuccessfulRequests
		row.GrossCostMicros = int64(selectedSuccessful) * sku.PriceMicrosPer1000 / 1000
		result[operation] = row
	}
	return result
}
