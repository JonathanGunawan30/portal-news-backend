package validation

import "testing"

type testContentPayload struct {
	Image  string `json:"image" validate:"required,safe_url"`
	Status string `json:"status" validate:"omitempty,oneof=PUBLISH DRAFT"`
}

func TestSafeURLAndStatusValidation(t *testing.T) {
	valid := testContentPayload{
		Image:  "https://cdn.example.com/image.jpg",
		Status: "PUBLISH",
	}
	if err := ValidateStruct(valid); err != nil {
		t.Fatalf("expected valid payload to pass, got: %v", err)
	}

	xssPayload := testContentPayload{
		Image:  "javascript:alert(1)",
		Status: "DRAFT",
	}
	if err := ValidateStruct(xssPayload); err == nil {
		t.Fatalf("expected javascript: URL to fail validation")
	}

	badStatus := testContentPayload{
		Image:  "https://cdn.example.com/image.jpg",
		Status: "INVALID_STATUS",
	}
	if err := ValidateStruct(badStatus); err == nil {
		t.Fatalf("expected invalid status to fail validation")
	}
}
