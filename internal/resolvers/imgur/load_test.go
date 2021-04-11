package imgur

import (
	"net/url"
	"testing"
	"time"

	"github.com/Chatterino/api/internal/mocks"
	qt "github.com/frankban/quicktest"
	"github.com/golang/mock/gomock"
	"github.com/koffeinsource/go-imgur"
)

func testLoadAndUnescape(c *qt.C, urlString string) (cleanTooltip string) {
	iret, _, err := load(urlString, nil)

	c.Assert(err, qt.IsNil)
	c.Assert(iret, qt.Not(qt.IsNil))

	response := iret.(response)

	c.Assert(response, qt.Not(qt.IsNil))
	c.Assert(response.err, qt.IsNil)

	c.Assert(response.resolverResponse, qt.Not(qt.IsNil))

	cleanTooltip, unescapeErr := url.PathUnescape(response.resolverResponse.Tooltip)
	c.Assert(unescapeErr, qt.IsNil)

	return cleanTooltip
}

func TestLoad(t *testing.T) {
	datetime := time.Date(2019, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()
	c := qt.New(t)
	mockCtrl := gomock.NewController(c)
	m := mocks.NewMockImgurClient(mockCtrl)
	apiClient = m

	c.Run("Normal image", func(c *qt.C) {
		const url = "image"
		m.
			EXPECT().
			GetInfoFromURL(gomock.Eq(url)).
			Return(&imgur.GenericInfo{
				Image: &imgur.ImageInfo{
					Title:       "My Cool Title",
					Description: "My Cool Description",
					Datetime:    int(datetime),
				},
				Album:  nil,
				GImage: nil,
				GAlbum: nil,
				Limit:  &imgur.RateLimit{},
			}, 420, nil)

		const expectedTooltip = `<div style="text-align: left;"><li><b>Title:</b> My Cool Title</li><li><b>Description:</b> My Cool Description</li><li><b>Uploaded:</b> 10 Nov 2019 • 23:00 UTC</li></div>`

		cleanTooltip := testLoadAndUnescape(c, url)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("NSFW image", func(c *qt.C) {
		const url = "nsfw_image"
		m.
			EXPECT().
			GetInfoFromURL(gomock.Eq(url)).
			Return(&imgur.GenericInfo{
				Image: &imgur.ImageInfo{
					Title:       "My Cool Title",
					Description: "My Cool Description",
					Datetime:    int(datetime),
					Nsfw:        true,
				},
				Album:  nil,
				GImage: nil,
				GAlbum: nil,
				Limit:  &imgur.RateLimit{},
			}, 420, nil)

		const expectedTooltip = `<div style="text-align: left;"><li><b>Title:</b> My Cool Title</li><li><b>Description:</b> My Cool Description</li><li><b>Uploaded:</b> 10 Nov 2019 • 23:00 UTC</li><li><b><span style="color: red;">NSFW</span></b></li></div>`

		cleanTooltip := testLoadAndUnescape(c, url)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("HTML tags in image", func(c *qt.C) {
		const url = "html_image"
		m.
			EXPECT().
			GetInfoFromURL(gomock.Eq(url)).
			Return(&imgur.GenericInfo{
				Image: &imgur.ImageInfo{
					Title:       "My <b>Cool</b> Title",
					Description: "My <b>Cool</b> Description",
					Datetime:    int(datetime),
				},
				Album:  nil,
				GImage: nil,
				GAlbum: nil,
				Limit:  &imgur.RateLimit{},
			}, 420, nil)

		const expectedTooltip = `<div style="text-align: left;"><li><b>Title:</b> My &lt;b&gt;Cool&lt;/b&gt; Title</li><li><b>Description:</b> My &lt;b&gt;Cool&lt;/b&gt; Description</li><li><b>Uploaded:</b> 10 Nov 2019 • 23:00 UTC</li></div>`

		cleanTooltip := testLoadAndUnescape(c, url)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("Empty album", func(c *qt.C) {
		const url = "empty_album"
		m.
			EXPECT().
			GetInfoFromURL(gomock.Eq(url)).
			Return(&imgur.GenericInfo{
				Image: nil,
				Album: &imgur.AlbumInfo{
					Title:       "Album Title",
					Description: "Album Description",
					DateTime:    int(datetime),
					ImagesCount: 0,
					Images:      []imgur.ImageInfo{},
				},
				GImage: nil,
				GAlbum: nil,
				Limit:  &imgur.RateLimit{},
			}, 420, nil)

		const expectedTooltip = `Empty album`

		cleanTooltip := testLoadAndUnescape(c, url)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("Normal album", func(c *qt.C) {
		const url = "album"
		m.
			EXPECT().
			GetInfoFromURL(gomock.Eq(url)).
			Return(&imgur.GenericInfo{
				Image: nil,
				Album: &imgur.AlbumInfo{
					Title:       "Album Title",
					Description: "Album Description",
					DateTime:    int(datetime),
					ImagesCount: 1,
					Images: []imgur.ImageInfo{
						{
							Title:       "My Cool Title",
							Description: "My Cool Description",
							Datetime:    int(datetime),
						},
					},
				},
				GImage: nil,
				GAlbum: nil,
				Limit:  &imgur.RateLimit{},
			}, 420, nil)

		const expectedTooltip = `<div style="text-align: left;"><li><b>Title:</b> Album Title</li><li><b>Description:</b> Album Description</li><li><b>Uploaded:</b> 10 Nov 2019 • 23:00 UTC</li></div>`

		cleanTooltip := testLoadAndUnescape(c, url)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("HTML tags in album", func(c *qt.C) {
		const url = "album"
		m.
			EXPECT().
			GetInfoFromURL(gomock.Eq(url)).
			Return(&imgur.GenericInfo{
				Image: nil,
				Album: &imgur.AlbumInfo{
					Title:       "Album <b>Title</b>",
					Description: "Album <b>Description</b>",
					DateTime:    int(datetime),
					ImagesCount: 1,
					Images: []imgur.ImageInfo{
						{
							Title:       "My Cool Title",
							Description: "My Cool Description",
							Datetime:    int(datetime),
						},
					},
				},
				GImage: nil,
				GAlbum: nil,
				Limit:  &imgur.RateLimit{},
			}, 420, nil)

		const expectedTooltip = `<div style="text-align: left;"><li><b>Title:</b> Album &lt;b&gt;Title&lt;/b&gt;</li><li><b>Description:</b> Album &lt;b&gt;Description&lt;/b&gt;</li><li><b>Uploaded:</b> 10 Nov 2019 • 23:00 UTC</li></div>`

		cleanTooltip := testLoadAndUnescape(c, url)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("Empty gallery album", func(c *qt.C) {
		const url = "empty_gallery_album"
		m.
			EXPECT().
			GetInfoFromURL(gomock.Eq(url)).
			Return(&imgur.GenericInfo{
				Image:  nil,
				Album:  nil,
				GImage: nil,
				GAlbum: &imgur.GalleryAlbumInfo{
					Title:       "Album Title",
					Description: "Album Description",
					ImagesCount: 0,
					Images:      []imgur.ImageInfo{},
				},
				Limit: &imgur.RateLimit{},
			}, 420, nil)

		const expectedTooltip = `Empty album`

		cleanTooltip := testLoadAndUnescape(c, url)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("Gallery album", func(c *qt.C) {
		const url = "empty_gallery_album"
		m.
			EXPECT().
			GetInfoFromURL(gomock.Eq(url)).
			Return(&imgur.GenericInfo{
				Image:  nil,
				Album:  nil,
				GImage: nil,
				GAlbum: &imgur.GalleryAlbumInfo{
					Title:       "Album Title",
					Description: "Album Description",
					DateTime:    int(datetime),
					ImagesCount: 1,
					Images: []imgur.ImageInfo{
						{
							Title:       "My Cool Title",
							Description: "My Cool Description",
							Datetime:    int(datetime),
						},
					},
				},
				Limit: &imgur.RateLimit{},
			}, 420, nil)

		const expectedTooltip = `<div style="text-align: left;"><li><b>Title:</b> Album Title</li><li><b>Description:</b> Album Description</li><li><b>Uploaded:</b> 10 Nov 2019 • 23:00 UTC</li></div>`

		cleanTooltip := testLoadAndUnescape(c, url)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("Gallery HTML tags album", func(c *qt.C) {
		const url = "empty_gallery_album"
		m.
			EXPECT().
			GetInfoFromURL(gomock.Eq(url)).
			Return(&imgur.GenericInfo{
				Image:  nil,
				Album:  nil,
				GImage: nil,
				GAlbum: &imgur.GalleryAlbumInfo{
					Title:       "Album <b>Title</b>",
					Description: "Album <b>Description</b>",
					DateTime:    int(datetime),
					ImagesCount: 1,
					Images: []imgur.ImageInfo{
						{
							Title:       "My Cool Title",
							Description: "My Cool Description",
							Datetime:    int(datetime),
						},
					},
				},
				Limit: &imgur.RateLimit{},
			}, 420, nil)

		const expectedTooltip = `<div style="text-align: left;"><li><b>Title:</b> Album &lt;b&gt;Title&lt;/b&gt;</li><li><b>Description:</b> Album &lt;b&gt;Description&lt;/b&gt;</li><li><b>Uploaded:</b> 10 Nov 2019 • 23:00 UTC</li></div>`

		cleanTooltip := testLoadAndUnescape(c, url)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	// TODO: Convert to Gallery images
	c.Run("Gallery image", func(c *qt.C) {
		const url = "gallery_image"
		m.
			EXPECT().
			GetInfoFromURL(gomock.Eq(url)).
			Return(&imgur.GenericInfo{
				Image: nil,
				Album: nil,
				GImage: &imgur.GalleryImageInfo{
					Title:       "My Cool Title",
					Description: "My Cool Description",
					Datetime:    int(datetime),
				},
				GAlbum: nil,
				Limit:  &imgur.RateLimit{},
			}, 420, nil)

		const expectedTooltip = `<div style="text-align: left;"><li><b>Title:</b> My Cool Title</li><li><b>Description:</b> My Cool Description</li><li><b>Uploaded:</b> 10 Nov 2019 • 23:00 UTC</li></div>`

		cleanTooltip := testLoadAndUnescape(c, url)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("NSFW gallery image", func(c *qt.C) {
		const url = "nsfw_gallery_image"
		m.
			EXPECT().
			GetInfoFromURL(gomock.Eq(url)).
			Return(&imgur.GenericInfo{
				Image: nil,
				Album: nil,
				GImage: &imgur.GalleryImageInfo{
					Title:       "My Cool Title",
					Description: "My Cool Description",
					Datetime:    int(datetime),
					Nsfw:        true,
				},
				GAlbum: nil,
				Limit:  &imgur.RateLimit{},
			}, 420, nil)

		const expectedTooltip = `<div style="text-align: left;"><li><b>Title:</b> My Cool Title</li><li><b>Description:</b> My Cool Description</li><li><b>Uploaded:</b> 10 Nov 2019 • 23:00 UTC</li><li><b><span style="color: red;">NSFW</span></b></li></div>`

		cleanTooltip := testLoadAndUnescape(c, url)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("HTML tags in gallery image", func(c *qt.C) {
		const url = "html_gallery_image"
		m.
			EXPECT().
			GetInfoFromURL(gomock.Eq(url)).
			Return(&imgur.GenericInfo{
				Image: nil,
				Album: nil,
				GImage: &imgur.GalleryImageInfo{
					Title:       "My <b>Cool</b> Title",
					Description: "My <b>Cool</b> Description",
					Datetime:    int(datetime),
				},
				GAlbum: nil,
				Limit:  &imgur.RateLimit{},
			}, 420, nil)

		const expectedTooltip = `<div style="text-align: left;"><li><b>Title:</b> My &lt;b&gt;Cool&lt;/b&gt; Title</li><li><b>Description:</b> My &lt;b&gt;Cool&lt;/b&gt; Description</li><li><b>Uploaded:</b> 10 Nov 2019 • 23:00 UTC</li></div>`

		cleanTooltip := testLoadAndUnescape(c, url)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})
}
