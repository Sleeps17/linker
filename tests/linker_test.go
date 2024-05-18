package tests

import (
	linkerV2 "github.com/Sleeps17/linker-protos/gen/go/linker"
	"github.com/Sleeps17/linker/pkg/random"
	"github.com/Sleeps17/linker/tests/suite"
	"github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"slices"
	"testing"
)

func TestLinkerPostAndDeleteTopic(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name      string
		username  string
		topic     string
		expectErr bool
	}{
		{
			name:      "Simple test 1",
			username:  generateUsername(),
			topic:     gofakeit.Word(),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := st.LinkerClient.PostTopic(ctx, &linkerV2.PostTopicRequest{
				Username: tt.username,
				Topic:    tt.topic,
			})

			require.Equal(t, tt.expectErr, err != nil)

			assert.NotEmpty(t, resp.GetTopicId())

			insertedId := resp.GetTopicId()

			_, err = st.LinkerClient.DeleteTopic(ctx, &linkerV2.DeleteTopicRequest{
				Username: tt.username,
				Topic:    tt.topic,
			})

			require.NoError(t, err)
			assert.Equal(t, insertedId, resp.GetTopicId())
		})
	}
}

func TestLinkerPostAndDeleteLink(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name      string
		username  string
		topic     string
		link      string
		alias     string
		expectErr bool
	}{
		{
			name:      "Simple test 1",
			username:  generateUsername(),
			topic:     gofakeit.Word(),
			link:      gofakeit.URL(),
			alias:     gofakeit.Word(),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.LinkerClient.PostTopic(ctx, &linkerV2.PostTopicRequest{
				Username: tt.username,
				Topic:    tt.topic,
			})

			require.NoError(t, err)

			resp, err := st.LinkerClient.PostLink(ctx, &linkerV2.PostLinkRequest{
				Username: tt.username,
				Topic:    tt.topic,
				Link:     tt.link,
				Alias:    tt.alias,
			})

			require.Equal(t, tt.expectErr, err != nil)

			if tt.alias != "" {
				assert.Equal(t, tt.alias, resp.Alias)
			} else {
				assert.NotEmpty(t, resp.GetAlias())
			}

			_, err = st.LinkerClient.DeleteLink(ctx, &linkerV2.DeleteLinkRequest{
				Username: tt.username,
				Topic:    tt.topic,
				Alias:    resp.GetAlias(),
			})

			require.NoError(t, err)
		})
	}
}

func TestLinkerPickLink(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name      string
		username  string
		topic     string
		link      string
		alias     string
		expectErr bool
	}{
		{
			name:      "Simple test 1",
			username:  generateUsername(),
			topic:     gofakeit.Word(),
			link:      gofakeit.URL(),
			alias:     gofakeit.Word(),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, err := st.LinkerClient.PostTopic(ctx, &linkerV2.PostTopicRequest{
				Username: tt.username,
				Topic:    tt.topic,
			})

			require.NoError(t, err)

			_, err = st.LinkerClient.PostLink(ctx, &linkerV2.PostLinkRequest{
				Username: tt.username,
				Topic:    tt.topic,
				Link:     tt.link,
				Alias:    tt.alias,
			})

			require.NoError(t, err)

			resp, err := st.LinkerClient.PickLink(ctx, &linkerV2.PickLinkRequest{
				Username: tt.username,
				Topic:    tt.topic,
				Alias:    tt.alias,
			})

			if tt.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.link, resp.Link)

			_, err = st.LinkerClient.DeleteLink(ctx, &linkerV2.DeleteLinkRequest{
				Username: tt.username,
				Topic:    tt.topic,
				Alias:    tt.alias,
			})

			require.NoError(t, err)
		})
	}
}

func generateLinks(length int) []string {
	res := make([]string, length)

	for i := range res {
		url := gofakeit.URL()

		for slices.Contains(res, url) {
			url = gofakeit.URL()
		}

		res[i] = url
	}

	return res
}

func generateAliases(length int) []string {
	res := make([]string, length)

	for i := range res {
		word := random.Alias()

		for slices.Contains(res, word) {
			word = random.Alias()
		}

		res[i] = word
	}

	return res
}

func generateUsername() string {
	var username string

	for len(username) < 8 {
		username = gofakeit.Username()
	}

	return username
}

func removeElement[T comparable](slice []T, element T) []T {
	result := make([]T, 0, len(slice))
	for _, value := range slice {
		if value != element {
			result = append(result, value)
		}
	}
	return result
}
