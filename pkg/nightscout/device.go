package nightscout

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/rest"
)

type DeviceGetter interface {
	DeviceStatus() DeviceInterface
}

type DeviceInterface interface {
	List(ctx context.Context, opts ListOptions) (*DeviceStatuses, error)
}

type device struct {
	client rest.Interface
}

// newGlucose returns a Glucos
func newDevice(c Client) *device {
	return &device{
		client: c.RESTClient(),
	}
}

func (d device) List(ctx context.Context, opts ListOptions) (result *DeviceStatuses, err error) {
	result = &DeviceStatuses{}
	r := d.client.Get().
		Name("devicestatus").
		Param("find[created_at][$gte]", opts.DateFrom.UTC().Format(time.RFC3339)).
		Param("find[created_at][$lte]", opts.DateTo.UTC().Format(time.RFC3339)).
		Param("count", strconv.Itoa(opts.Count))

	err = r.Do(ctx).Into(result)

	if len(opts.Kind) > 0 {
		result = result.Filter(OnlyDeviceType(opts.Kind))
	}

	return
}

type Uploader struct {
	Type    string `json:"type"`
	Battery int    `json:"battery"`
}

type DeviceStatus struct {
	ID        string    `json:"_id"`
	Device    string    `json:"device"`
	CreatedAt time.Time `json:"created_at"`
	UtcOffset int       `json:"utcOffset"`
	Uploader  *Uploader `json:"uploader"`
}

func (d *DeviceStatus) Kind() string {
	return "DeviceStatus"
}

type DeviceStatuses []*DeviceStatus

func (r *DeviceStatuses) Append(e *DeviceStatus) {
	*r = append(*r, e)
}

func (r DeviceStatuses) Len() int {
	return len(r)
}

type DeviceStatusFilterFunc func(*DeviceStatus) bool

func OnlyDeviceType(deviceType string) DeviceStatusFilterFunc {
	return func(d *DeviceStatus) bool {
		return strings.ToUpper(d.Uploader.Type) == strings.ToUpper(deviceType)
	}
}

func (ds *DeviceStatuses) Filter(fn DeviceStatusFilterFunc) (result *DeviceStatuses) {
	result = &DeviceStatuses{}
	ds.Visit(func(e *DeviceStatus, _ error) error {
		if fn(e) {
			result.Append(e)
		}
		return nil
	})
	return result
}

type DeviceStatusVisitorFunc func(*DeviceStatus, error) error

func (ds DeviceStatuses) Visit(fn DeviceStatusVisitorFunc) error {
	var err error
	for _, entry := range ds {
		if err = fn(entry, err); err != nil {
			return err
		}
	}
	return nil
}
