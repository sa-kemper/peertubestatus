package peertubeApi

import (
	"net/url"
	"reflect"
	"testing"
)

func Test_toQueryParams(t *testing.T) {
	// User represents a user profile in the system
	type User struct {
		ID          int64    `json:"id"`
		Username    string   `json:"username"`
		Email       string   `json:"email"`
		Age         int      `json:"age"`
		Roles       []string `json:"roles"`
		IsActive    bool     `json:"is_active"`
		Permissions []int    `json:"permissions"`
	}

	// Product represents an inventory item
	type Product struct {
		ProductID     int64    `json:"product_id"`
		Name          string   `json:"name"`
		Description   string   `json:"description"`
		Price         float64  `json:"price"`
		StockQuantity int      `json:"stock_quantity"`
		Categories    []string `json:"categories"`
		IsAvailable   bool     `json:"is_available"`
	}

	// Location represents geographical information
	type Location struct {
		LocationID  int64    `json:"location_id"`
		Name        string   `json:"name"`
		Latitude    float64  `json:"latitude"`
		Longitude   float64  `json:"longitude"`
		Country     string   `json:"country"`
		Regions     []string `json:"regions"`
		IsPopulated bool     `json:"is_populated"`
	}

	// Book represents a literary work
	type Book struct {
		ISBN        string   `json:"isbn"`
		Title       string   `json:"title"`
		Author      string   `json:"author"`
		PublishYear int      `json:"publish_year"`
		Genres      []string `json:"genres"`
		PageCount   int      `json:"page_count"`
		IsAvailable bool     `json:"is_available"`
	}

	type args struct {
		paramObject any
	}
	tests := []struct {
		name string
		args args
		want url.Values
	}{
		{
			name: "User Struct Query Params",
			args: args{
				paramObject: User{
					ID:          12345,
					Username:    "john_doe",
					Email:       "john@example.com",
					Age:         30,
					Roles:       []string{"admin", "user"},
					IsActive:    true,
					Permissions: []int{1, 2, 3},
				},
			},
			want: url.Values{
				"iD":          []string{"12345"},
				"username":    []string{"john_doe"},
				"email":       []string{"john@example.com"},
				"age":         []string{"30"},
				"roles":       []string{"admin", "user"},
				"isActive":    []string{"true"},
				"permissions": []string{"1", "2", "3"},
			},
		},
		{
			name: "Product Struct Query Params",
			args: args{
				paramObject: Product{
					ProductID:     6789,
					Name:          "Smart Widget",
					Description:   "Advanced technological device",
					Price:         99.99,
					StockQuantity: 50,
					Categories:    []string{"electronics", "gadgets"},
					IsAvailable:   true,
				},
			},
			want: url.Values{
				"productID":     []string{"6789"},
				"name":          []string{"Smart Widget"},
				"description":   []string{"Advanced technological device"},
				"price":         []string{"99.99"},
				"stockQuantity": []string{"50"},
				"categories":    []string{"electronics", "gadgets"},
				"isAvailable":   []string{"true"},
			},
		},
		{
			name: "Location Struct Query Params",
			args: args{
				paramObject: Location{
					LocationID:  101,
					Name:        "Mountain View",
					Latitude:    37.4220,
					Longitude:   -122.0841,
					Country:     "United States",
					Regions:     []string{"California", "Bay Area"},
					IsPopulated: true,
				},
			},
			want: url.Values{
				"locationID":  []string{"101"},
				"name":        []string{"Mountain View"},
				"latitude":    []string{"37.422"},
				"longitude":   []string{"-122.0841"},
				"country":     []string{"United States"},
				"regions":     []string{"California", "Bay Area"},
				"isPopulated": []string{"true"},
			},
		},
		{
			name: "Book Struct Query Params",
			args: args{
				paramObject: Book{
					ISBN:        "978-3-16-148410-0",
					Title:       "The Great Novel",
					Author:      "Jane Doe",
					PublishYear: 2023,
					Genres:      []string{"Fiction", "Drama"},
					PageCount:   350,
					IsAvailable: true,
				},
			},
			want: url.Values{
				"iSBN":        []string{"978-3-16-148410-0"},
				"title":       []string{"The Great Novel"},
				"author":      []string{"Jane Doe"},
				"publishYear": []string{"2023"},
				"genres":      []string{"Fiction", "Drama"},
				"pageCount":   []string{"350"},
				"isAvailable": []string{"true"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toQueryParams(tt.args.paramObject); !reflect.DeepEqual(got, tt.want) {
				for i, val := range got {
					if !reflect.DeepEqual(got[i], tt.want[i]) {
						t.Errorf("toQueryParams()[%s] = %v, want %v", i, val, tt.want[i])
					}
				}
			}
		})
	}
}
