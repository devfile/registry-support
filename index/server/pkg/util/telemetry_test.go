package util

import (
	"github.com/gin-gonic/gin"
	"github.com/kylelemons/godebug/pretty"
	"gopkg.in/segmentio/analytics-go.v3"
	"net"
	"net/http"
	"reflect"
	"testing"
)

func TestGetUser(t *testing.T) {
	tests := []struct {
		name    string
		context *gin.Context
		want    string
	}{
		{
			name: "User header is set",
			context: &gin.Context{
				Request: &http.Request{
					Header: http.Header{
						"User": {"testuser"},
					},
				},
			},
			want: "testuser",
		},
		{
			name: "User header is unset",
			context: &gin.Context{
				Request: &http.Request{
					Header: http.Header{},
				},
			},
			want: defaultUser,
		},
		{
			name: "Multiple users set",
			context: &gin.Context{
				Request: &http.Request{
					Header: http.Header{
						"User": {"user1", "user2", "user3"},
					},
				},
			},
			want: "user1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user := GetUser(test.context)
			if user != test.want {
				t.Errorf("Got: %v, Expected: %v", user, test.want)
			}
		})
	}
}

func TestSetContext(t *testing.T) {
	tests := []struct {
		name    string
		context *gin.Context
		want    *analytics.Context
	}{
		{
			name: "Accept-Language is set with region",
			context: &gin.Context{
				Request: &http.Request{
					Header: http.Header{
						"Accept-Language": {"en_CA"},
					},
				},
			},
			want: &analytics.Context{
				Location: analytics.LocationInfo{
					Country: "CA",
				},
				IP: net.IPv4(0, 0, 0, 0),
			},
		},
		{
			name: "Accept-Language is set with no region",
			context: &gin.Context{
				Request: &http.Request{
					Header: http.Header{
						"Accept-Language": {"en"},
					},
				},
			},
			want: &analytics.Context{
				Location: analytics.LocationInfo{
					Country: "US",
				},
				IP: net.IPv4(0, 0, 0, 0),
			},
		},
		{
			name: "Accept-Language is unset",
			context: &gin.Context{
				Request: &http.Request{
					Header: http.Header{},
				},
			},
			want: &analytics.Context{
				Location: analytics.LocationInfo{
					Country: "",
				},
				IP: net.IPv4(0, 0, 0, 0),
			},
		},
		{
			name: "Accept-Language has a weighted list",
			context: &gin.Context{
				Request: &http.Request{
					Header: http.Header{
						"Accept-Language": {"gsw", "en;q=0.7", "en-US;q=0.8"},
					},
				},
			},
			want: &analytics.Context{
				Location: analytics.LocationInfo{
					Country: "CH",
				},
				IP: net.IPv4(0, 0, 0, 0),
			},
		},
		{
			name: "Accept-Language has a wildcard",
			context: &gin.Context{
				Request: &http.Request{
					Header: http.Header{
						"Accept-Language": {"*"},
					},
				},
			},
			want: &analytics.Context{
				Location: analytics.LocationInfo{
					Country: "",
				},
				IP: net.IPv4(0, 0, 0, 0),
			},
		},
		{
			name: "Accept-Language has an invalid locale",
			context: &gin.Context{
				Request: &http.Request{
					Header: http.Header{
						"Accept-Language": {"invalid"},
					},
				},
			},
			want: &analytics.Context{
				Location: analytics.LocationInfo{
					Country: "",
				},
				IP: net.IPv4(0, 0, 0, 0),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context := SetContext(test.context)
			if !reflect.DeepEqual(context, test.want) {
				t.Errorf("Got: %v, Expected: %v.  Differences are %v ", context, test.want, pretty.Compare(context, test.want))
			}
		})
	}
}
