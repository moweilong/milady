// Package copier is github.com/jinzhu/copier,  default option is add converters for time.Time <--> String
package copier

import (
	"fmt"
	"time"

	"github.com/jinzhu/copier"
)

// Copy src to dst with option Converters for time.Time <--> String
func Copy(dst interface{}, src interface{}) error {
	return copier.CopyWithOption(dst, src, Converter)
}

// Converter Converters for time.Time <--> String
var Converter = copier.Option{
	DeepCopy: true,
	Converters: []copier.TypeConverter{
		// time.Time to string
		{
			SrcType: time.Time{},
			DstType: copier.String,
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(time.Time)
				if !ok {
					return nil, fmt.Errorf("expected time.Time got %T", src)
				}
				return s.Format(time.RFC3339), nil
			},
		},

		// *time.Time to string
		{
			SrcType: &time.Time{},
			DstType: copier.String,
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(*time.Time)
				if !ok {
					return nil, fmt.Errorf("expected *time.Time got %T", src)
				}
				if s == nil {
					return "", nil
				}
				return s.Format(time.RFC3339), nil
			},
		},

		// string to time.Time
		{
			SrcType: copier.String,
			DstType: time.Time{},
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(string)
				if !ok {
					return nil, fmt.Errorf("expected string got %T", src)
				}
				if s == "" {
					return time.Time{}, nil
				}
				return time.Parse(time.RFC3339, s)
			},
		},

		// string to *time.Time
		{
			SrcType: copier.String,
			DstType: &time.Time{},
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(string)
				if !ok {
					return nil, fmt.Errorf("expected string got %T", src)
				}
				if s == "" {
					return nil, nil
				}
				t, err := time.Parse(time.RFC3339, s)
				return &t, err
			},
		},
	},
}

// CopyDefault copy src to dst with default option
func CopyDefault(dst interface{}, src interface{}) error {
	return copier.Copy(dst, src)
}

// CopyWithOption copy src to dst with option
func CopyWithOption(dst interface{}, src interface{}, options copier.Option) error {
	return copier.CopyWithOption(dst, src, options)
}
