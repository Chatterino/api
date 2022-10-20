package utils

import "github.com/davidbyttow/govips/v2/vips"

// MimeType turn a vips.ImageType into a MIME type string
func MimeType(imgType vips.ImageType) string {
	subtype, ok := vips.ImageTypes[imgType]
	if ok {
		return "image/" + subtype
	}

	return ""
}
