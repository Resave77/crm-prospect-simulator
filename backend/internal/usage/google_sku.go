package usage

import "strings"

// ResolveGoogleSKU applies compatibility rules without changing the stored
// ProviderUsageEvent. The returned operation is the verified pricing key.
func ResolveGoogleSKU(operation, skuCategory, fieldMask string) (string, GoogleSKU, bool) {
	if skuCategory != "" {
		if sku, ok := ClassifyGoogleSKU(skuCategory, fieldMask); ok {
			return skuCategory, sku, true
		}
	}
	if sku, ok := ClassifyGoogleSKU(operation, fieldMask); ok {
		return operation, sku, true
	}
	if operation == "PLACE_DETAILS" {
		for _, candidate := range []string{"BUSINESS_INFO", "CORE_DETAIL"} {
			if sku, ok := ClassifyGoogleSKU(candidate, fieldMask); ok {
				return candidate, sku, true
			}
		}
	}
	return "", GoogleSKU{}, false
}

// ClassifyGoogleSKU is deterministic and field-mask driven. Unknown masks are
// intentionally not guessed into a billable tier.
func ClassifyGoogleSKU(operation, fieldMask string) (GoogleSKU, bool) {
	p := DefaultPricing()
	if operation == "NEARBY_SEARCH" || operation == "TEXT_SEARCH" {
		sku, ok := p.GoogleSKUs[operation]
		return sku, ok && strings.Contains(fieldMask, "displayName")
	}
	if operation == "CORE_DETAIL" {
		sku, ok := p.GoogleSKUs["CORE_DETAIL"]
		return sku, ok && strings.Contains(fieldMask, "displayName") && strings.Contains(fieldMask, "googleMapsUri")
	}
	if operation == "BUSINESS_INFO" {
		sku, ok := p.GoogleSKUs["BUSINESS_INFO"]
		return sku, ok && strings.Contains(fieldMask, "nationalPhoneNumber") && strings.Contains(fieldMask, "regularOpeningHours")
	}
	if operation == "PLACE_PHOTO" {
		sku, ok := p.GoogleSKUs["PLACE_PHOTO"]
		return sku, ok
	}
	return GoogleSKU{}, false
}
