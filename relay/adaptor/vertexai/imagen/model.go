package imagen

// CreateImageRequest is the request body for the Imagen API.
//
// https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/imagen-api
type CreateImageRequest struct {
	Instances  []createImageInstance `json:"instances" binding:"required,min=1"`
	Parameters createImageParameters `json:"parameters" binding:"required"`
}

type createImageInstance struct {
	Prompt          string           `json:"prompt"`
	ReferenceImages []ReferenceImage `json:"referenceImages,omitempty"`
	Image           *promptImage     `json:"image,omitempty"` // Keeping for backward compatibility
}

// ReferenceImage represents a reference image for the Imagen edit API
//
// https://console.cloud.google.com/vertex-ai/publishers/google/model-garden/imagen-3.0-capability-001?project=ai-ca-447000
type ReferenceImage struct {
	ReferenceType  string             `json:"referenceType" binding:"required,oneof=REFERENCE_TYPE_RAW REFERENCE_TYPE_MASK"`
	ReferenceId    int                `json:"referenceId"`
	ReferenceImage ReferenceImageData `json:"referenceImage"`
	// MaskImageConfig is used when ReferenceType is "REFERENCE_TYPE_MASK",
	// to provide a mask image for the reference image.
	MaskImageConfig *MaskImageConfig `json:"maskImageConfig,omitempty"`
}

// ReferenceImageData contains the actual image data
type ReferenceImageData struct {
	BytesBase64Encoded string  `json:"bytesBase64Encoded,omitempty"`
	GcsUri             *string `json:"gcsUri,omitempty"`
	MimeType           *string `json:"mimeType,omitempty" binding:"omitempty,oneof=image/jpeg image/png"`
}

// MaskImageConfig specifies how to use the mask image
type MaskImageConfig struct {
	// MaskMode is used to mask mode for mask editing.
	// Set MASK_MODE_USER_PROVIDED for input user provided mask in the B64_MASK_IMAGE,
	// MASK_MODE_BACKGROUND for automatically mask out background without user provided mask,
	// MASK_MODE_SEMANTIC for automatically generating semantic object masks by
	// specifying a list of object class IDs in maskClasses.
	MaskMode    string   `json:"maskMode" binding:"required,oneof=MASK_MODE_USER_PROVIDED MASK_MODE_BACKGROUND MASK_MODE_SEMANTIC"`
	MaskClasses []int    `json:"maskClasses,omitempty"` // Object class IDs when maskMode is MASK_MODE_SEMANTIC
	Dilation    *float64 `json:"dilation,omitempty"`    // Determines the dilation percentage of the mask provided. Min: 0, Max: 1, Default: 0.03
}

// promptImage is the image to be used as a prompt for the Imagen API.
// It can be either a base64 encoded image or a GCS URI.
type promptImage struct {
	BytesBase64Encoded *string `json:"bytesBase64Encoded,omitempty"`
	GcsUri             *string `json:"gcsUri,omitempty"`
	MimeType           *string `json:"mimeType,omitempty" binding:"omitempty,oneof=image/jpeg image/png"`
}

type createImageParameters struct {
	SampleCount int     `json:"sampleCount" binding:"required,min=1"`
	Mode        *string `json:"mode,omitempty" binding:"omitempty,oneof=upscaled"`
	// EditMode set edit mode for mask editing.
	// Set EDIT_MODE_INPAINT_REMOVAL for inpainting removal,
	// EDIT_MODE_INPAINT_INSERTION for inpainting insert,
	// EDIT_MODE_OUTPAINT for outpainting,
	// EDIT_MODE_BGSWAP for background swap.
	EditMode      *string        `json:"editMode,omitempty" binding:"omitempty,oneof=EDIT_MODE_INPAINT_REMOVAL EDIT_MODE_INPAINT_INSERTION EDIT_MODE_OUTPAINT EDIT_MODE_BGSWAP"`
	UpscaleConfig *upscaleConfig `json:"upscaleConfig,omitempty"`
	Seed          *int64         `json:"seed,omitempty"`
}

type upscaleConfig struct {
	// UpscaleFactor is the factor to which the image will be upscaled.
	// If not specified, the upscale factor will be determined from
	// the longer side of the input image and sampleImageSize.
	// Available values: x2 or x4 .
	UpscaleFactor *string `json:"upscaleFactor,omitempty" binding:"omitempty,oneof=2x 4x"`
}

// CreateImageResponse is the response body for the Imagen API.
type CreateImageResponse struct {
	Predictions []createImageResponsePrediction `json:"predictions"`
}

type createImageResponsePrediction struct {
	MimeType           string `json:"mimeType"`
	BytesBase64Encoded string `json:"bytesBase64Encoded"`
}

// VQARequest is the response body for the Visual Question Answering API.
//
// https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/visual-question-answering
type VQAResponse struct {
	Predictions []string `json:"predictions"`
}
