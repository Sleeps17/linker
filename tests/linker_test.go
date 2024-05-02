package tests

import (
	linkerV1 "github.com/Sleeps17/linker-protos/gen/go/linker"
	"github.com/Sleeps17/linker/tests/suite"
	"github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				return
			}

			if tt.alias != "" {
				assert.Equal(t, tt.alias, resp.Alias)
			} else {
				assert.NotEmpty(t, resp.GetAlias())
			}
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
			} else {
				require.Error(t, err)
				return
			}

			assert.Equal(t, tt.link, resp.Link)
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
			username:    generateUsername(),
			links:       generateLinks(20),
			aliases:     generateAliases(20),
			expectedErr: false,
		},
		{
			name:        "normal case 2",
			username:    generateUsername(),
			links:       generateLinks(20),
			aliases:     generateAliases(20),
			expectedErr: false,
		},
		{
			name:        "normal case 3",
			username:    generateUsername(),
			links:       generateLinks(20),
			aliases:     generateAliases(20),
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
		{
			name:        "with empty links and aliases",
			username:    generateUsername(),
			links:       nil,
			aliases:     nil,
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := range tt.links {
				_, _ = st.LinkerClient.Post(ctx, &linkerV1.PostRequest{
					Username: tt.username,
					Link:     tt.links[i],
					Alias:    tt.aliases[i],
				})

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

			assert.Equal(t, tt.aliases, resp.Links)
		})
	}
}

func TestLinkerDelete(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name     string
		username string
		link     string
		alias    string
	}{
		{},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, _ = st.LinkerClient.Post(ctx, &linkerV1.PostRequest{
				Username: tt.username,
				Link:     tt.link,
				Alias:    tt.alias,
			})

			_, _ = st.LinkerClient.Delete(ctx, &linkerV1.DeleteRequest{
				Username: tt.username,
				Alias:    tt.alias,
			})
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
		res[i] = gofakeit.URL()
	}

	return res
}

func generateAliases(length int) []string {
	res := make([]string, length)

	for i := range res {
		res[i] = gofakeit.Word()
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
