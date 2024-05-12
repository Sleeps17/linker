package tests

import (
	linkerV1 "github.com/Sleeps17/linker-protos/gen/go/linker"
	"github.com/Sleeps17/linker/pkg/random"
	"github.com/Sleeps17/linker/tests/suite"
	"github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"slices"
	"testing"
)

func TestLinkerPost(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		username    string
		link        string
		alias       string
		expectedErr bool
	}{
		{
			name:        "normal case 1",
			username:    generateUsername(),
			link:        gofakeit.URL(),
			alias:       gofakeit.Word(),
			expectedErr: false,
		},
		{
			name:        "normal case 2",
			username:    generateUsername(),
			link:        gofakeit.URL(),
			alias:       gofakeit.Word(),
			expectedErr: false,
		},
		{
			name:        "with empty alias",
			username:    generateUsername(),
			link:        gofakeit.URL(),
			alias:       "",
			expectedErr: false,
		},
		{
			name:        "with empty username",
			username:    "",
			link:        gofakeit.URL(),
			alias:       gofakeit.Word(),
			expectedErr: true,
		},
		{
			name:        "with username less then 8 characters",
			username:    generateUsername()[0:5],
			link:        gofakeit.URL(),
			alias:       gofakeit.Word(),
			expectedErr: true,
		},
		{
			name:        "with empty link",
			username:    generateUsername(),
			link:        "",
			alias:       gofakeit.Word(),
			expectedErr: true,
		},
		{
			name:        "with invalid link",
			username:    generateUsername(),
			link:        "invalid",
			alias:       gofakeit.Word(),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := st.LinkerClient.Post(ctx, &linkerV1.PostRequest{
				Username: tt.username,
				Link:     tt.link,
				Alias:    tt.alias,
			})

			if !tt.expectedErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				return
			}

			if tt.alias != "" {
				assert.Equal(t, tt.alias, resp.Alias)
			} else {
				assert.NotEmpty(t, resp.GetAlias())
			}

			_, err = st.LinkerClient.Delete(ctx, &linkerV1.DeleteRequest{
				Alias:    resp.GetAlias(),
				Username: tt.username,
			})

			assert.NoError(t, err)
		})
	}
}

func TestLinkerPick(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		username    string
		link        string
		alias       string
		changeAlias bool
		expectedErr bool
	}{
		{
			name:        "normal case 1",
			username:    generateUsername(),
			link:        gofakeit.URL(),
			alias:       gofakeit.Word(),
			changeAlias: false,
			expectedErr: false,
		},
		{
			name:        "normal case 2",
			username:    generateUsername(),
			link:        gofakeit.URL(),
			alias:       gofakeit.Word(),
			changeAlias: false,
			expectedErr: false,
		},
		{
			name:        "with empty alias",
			username:    generateUsername(),
			link:        gofakeit.URL(),
			alias:       "",
			changeAlias: false,
			expectedErr: true,
		},
		{
			name:        "with empty username",
			username:    "",
			link:        gofakeit.URL(),
			alias:       gofakeit.Word(),
			changeAlias: false,
			expectedErr: true,
		},
		{
			name:        "with invalid username",
			username:    generateUsername()[0:5],
			link:        gofakeit.URL(),
			alias:       gofakeit.Word(),
			changeAlias: false,
			expectedErr: true,
		},
		{
			name:        "with undefined alias",
			username:    generateUsername(),
			link:        gofakeit.URL(),
			alias:       "",
			changeAlias: true,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, _ = st.LinkerClient.Post(ctx, &linkerV1.PostRequest{
				Username: tt.username,
				Link:     tt.link,
				Alias:    tt.alias,
			})

			if tt.changeAlias {
				tt.alias = "some undefined alias"
			}

			resp, err := st.LinkerClient.Pick(ctx, &linkerV1.PickRequest{
				Username: tt.username,
				Alias:    tt.alias,
			})

			if !tt.expectedErr {
				require.NoError(t, err)
				assert.Equal(t, tt.link, resp.Link)

				_, err = st.LinkerClient.Delete(ctx, &linkerV1.DeleteRequest{
					Username: tt.username,
					Alias:    tt.alias,
				})

				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				return
			}
		})
	}
}

func TestLinkerList(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		username    string
		links       []string
		aliases     []string
		expectedErr bool
	}{
		{
			name:        "normal case 1",
			username:    random.Alias(rand.Int()%10 + 10),
			links:       generateLinks(8),
			aliases:     generateAliases(8),
			expectedErr: false,
		},
		{
			name:        "normal case 2",
			username:    random.Alias(rand.Int()%10 + 10),
			links:       generateLinks(8),
			aliases:     generateAliases(8),
			expectedErr: false,
		},
		{
			name:        "normal case 3",
			username:    random.Alias(rand.Int()%10 + 10),
			links:       generateLinks(8),
			aliases:     generateAliases(8),
			expectedErr: false,
		},
		{
			name:        "with empty username",
			username:    "",
			expectedErr: true,
		},
		{
			name:        "with invalid username",
			username:    generateUsername()[0:5],
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			indexForDelete := make([]int, 0, len(tt.links))
			for i := range tt.links {
				_, err := st.LinkerClient.Post(ctx, &linkerV1.PostRequest{
					Username: tt.username,
					Link:     tt.links[i],
					Alias:    tt.aliases[i],
				})

				if err != nil {
					indexForDelete = append(indexForDelete, i)
				}
			}

			cnt := 0
			for _, elem := range indexForDelete {
				tt.links = removeElement(tt.links, tt.links[elem-cnt])
				tt.aliases = removeElement(tt.aliases, tt.aliases[elem-cnt])
				cnt++
			}

			resp, err := st.LinkerClient.List(ctx, &linkerV1.ListRequest{
				Username: tt.username,
			})

			if !tt.expectedErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				return
			}

			slices.Sort(tt.aliases)
			slices.Sort(tt.links)
			slices.Sort(resp.Aliases)
			slices.Sort(resp.Links)

			assert.Equal(t, tt.links, resp.Links)
			assert.Equal(t, tt.aliases, resp.Aliases)

			for i := range tt.links {
				_, err := st.LinkerClient.Delete(ctx, &linkerV1.DeleteRequest{
					Username: tt.username,
					Alias:    tt.aliases[i],
				})

				assert.NoError(t, err)
			}
		})
	}
}

func TestLinkerDelete(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		username    string
		link        string
		alias       string
		expectedErr bool
	}{
		{
			name:     "normal case 1",
			username: generateUsername(),
			link:     gofakeit.URL(),
			alias:    gofakeit.Word(),
		},
		{
			name:     "normal case 1",
			username: generateUsername(),
			link:     gofakeit.URL(),
			alias:    gofakeit.Word(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, err := st.LinkerClient.Post(ctx, &linkerV1.PostRequest{
				Username: tt.username,
				Link:     tt.link,
				Alias:    tt.alias,
			})

			require.NoError(t, err)

			_, err = st.LinkerClient.Delete(ctx, &linkerV1.DeleteRequest{
				Username: tt.username,
				Alias:    tt.alias,
			})

			if !tt.expectedErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				return
			}
		})
	}
}

func TestLinkerPost_RecordAlreadyExists(t *testing.T) {
	ctx, st := suite.New(t)

	username := generateUsername()
	link := gofakeit.URL()
	alias := gofakeit.Word()

	_, err := st.LinkerClient.Post(ctx, &linkerV1.PostRequest{
		Username: username,
		Link:     link,
		Alias:    alias,
	})

	require.NoError(t, err)

	link = gofakeit.URL()

	_, err = st.LinkerClient.Post(ctx, &linkerV1.PostRequest{
		Username: username,
		Link:     link,
		Alias:    alias,
	})

	require.Error(t, err)

	_, err = st.LinkerClient.Delete(ctx, &linkerV1.DeleteRequest{
		Username: username,
		Alias:    alias,
	})

	require.NoError(t, err)
}

func TestLinkerPick_UnknownUsername(t *testing.T) {
	ctx, st := suite.New(t)

	username := generateUsername()
	alias := "some random word"

	resp, err := st.LinkerClient.Pick(ctx, &linkerV1.PickRequest{
		Username: username,
		Alias:    alias,
	})

	require.Error(t, err)
	require.Nil(t, resp)
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
