package imgur

import (
	"log"
	"net/url"
	"path"
	"strings"

	"github.com/Chatterino/api/pkg/humanize"
	"github.com/koffeinsource/go-imgur"
)

// Make a miniImage struct from a Gallery Image Info go-imgur struct
func makeMiniImageFromGImage(imageInfo imgur.GalleryImageInfo) miniImage {
	mini := miniImage{
		Title:       imageInfo.Title,
		Description: imageInfo.Description,
		UploadDate:  humanize.CreationDateTimeUnix(int64(imageInfo.Datetime)),
		Nsfw:        imageInfo.Nsfw,
		Animated:    imageInfo.Animated,
		Album:       true,

		mimeType: imageInfo.MimeType,
		size:     imageInfo.Size,

		Link: imageInfo.Link,
	}

	finalizeMiniImage(&mini)

	return mini
}

// Make a miniImage struct from an Image Info go-imgur struct
func makeMiniImage(imageInfo imgur.ImageInfo) miniImage {
	mini := miniImage{
		Title:       imageInfo.Title,
		Description: imageInfo.Description,
		UploadDate:  humanize.CreationDateTimeUnix(int64(imageInfo.Datetime)),
		Nsfw:        imageInfo.Nsfw,
		Animated:    imageInfo.Animated,
		Album:       true,

		mimeType: imageInfo.MimeType,
		size:     imageInfo.Size,

		Link: imageInfo.Link,
	}

	finalizeMiniImage(&mini)

	return mini
}

// Do some final work on a mini image
// If an image is animated, and it's not a straight up gif, make the thumbnail a static image
// If the image is not animated, limit the size of the thumbnail
// see top of resolver.go for the max image size before downscaling the thumbnail
func finalizeMiniImage(mini *miniImage) {
	if mini.Animated {

		if mini.mimeType != "image/gif" {
			// Animated image in an 'unthumbnailable' format
			// We try to get a thumbnail from the .png link of the same image
			if linkURL, err := url.Parse(mini.Link); err == nil {
				ext := path.Ext(linkURL.Path)
				linkURL.Path = strings.Replace(linkURL.Path, ext, ".png", 1)
				mini.Link = linkURL.String()
			} else {
				log.Println("[IMGUR] Error making static thumbnail for image:", err, mini)
				mini.Link = ""
			}
		}
	} else {
		if mini.size > maxRawImageSize {
			if linkURL, err := url.Parse(mini.Link); err == nil {
				ext := path.Ext(linkURL.Path)
				linkURL.Path = strings.Replace(linkURL.Path, ext, "l"+ext, 1)
				mini.Link = linkURL.String()
			} else {
				log.Println("[IMGUR] Error making smaller thumbnail for image:", err, mini)
			}
		}
	}

	if mini.Nsfw {
		// Hide thumbnails for NSFW images
		mini.Link = ""
	}
}
