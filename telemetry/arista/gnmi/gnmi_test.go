//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package gnmi

import (
	"bytes"
	"context"
	"testing"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/status"
	"github.com/yahoo/panoptes-stream/telemetry"
	"github.com/yahoo/panoptes-stream/telemetry/mock"
)

func TestAristaSimplePath(t *testing.T) {
	var (
		addr    = "127.0.0.1:50500"
		ch      = make(telemetry.ExtDSChan, 1)
		ctx     = context.Background()
		sensors []*config.Sensor
	)
	ln, err := mock.StartGNMIServer(addr, mock.Update{Notification: mock.AristaUpdate(), Attempt: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	cfg := config.NewMockConfig()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	sensors = append(sensors, &config.Sensor{
		Service: "arista.gnmi",
		Output:  "console::stdout",
		Path:    "/interfaces/interface/state/counters",
	})

	g := New(cfg.Logger(), conn, sensors, ch)
	g.Start(ctx)

	resp := <-ch

	assert.Equal(t, sensors[0].Path, resp.DS["prefix"].(string))
	assert.Equal(t, "127.0.0.1", resp.DS["system_id"].(string))
	assert.Equal(t, int64(1595363593437180059), resp.DS["timestamp"].(int64))
	assert.Equal(t, "Ethernet1", resp.DS["labels"].(map[string]string)["name"])
	assert.Equal(t, "out-octets", resp.DS["key"].(string))
	assert.Equal(t, int64(50302030597), resp.DS["value"].(int64))
	assert.Equal(t, "console::stdout", resp.Output)

	assert.Equal(t, "", cfg.LogOutput.String())
}

func TestAristaBGPSimplePath(t *testing.T) {
	var (
		addr    = "127.0.0.1:50500"
		ch      = make(telemetry.ExtDSChan, 1)
		ctx     = context.Background()
		sensors []*config.Sensor
	)
	ln, err := mock.StartGNMIServer(addr, mock.Update{Notification: mock.AristaBGPUpdate(), Attempt: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	cfg := config.NewMockConfig()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	sensors = append(sensors, &config.Sensor{
		Service: "arista.gnmi",
		Output:  "console::stdout",
		Path:    "/network-instances/network-instance",
	})

	g := New(cfg.Logger(), conn, sensors, ch)
	g.Start(ctx)

	resp := <-ch

	assert.Equal(t, sensors[0].Path, resp.DS["prefix"].(string))
	assert.Equal(t, "127.0.0.1", resp.DS["system_id"].(string))
	assert.Equal(t, int64(1595363593413814979), resp.DS["timestamp"].(int64))
	assert.Equal(t, "default", resp.DS["labels"].(map[string]string)["name"])
	assert.Equal(t, "BGP", resp.DS["labels"].(map[string]string)["identifier"])
	assert.Equal(t, "IPV6_UNICAST", resp.DS["labels"].(map[string]string)["afi-safi-name"])
	assert.Equal(t, "BGP", resp.DS["labels"].(map[string]string)["_name"])
	assert.Equal(t, "protocols/protocol/bgp/global/afi-safis/afi-safi/config/afi-safi-name", resp.DS["key"].(string))
	assert.Equal(t, "openconfig-bgp-types:IPV6_UNICAST", resp.DS["value"].(string))
	assert.Equal(t, "console::stdout", resp.Output)

	assert.Equal(t, "", cfg.LogOutput.String())
}

func TestAristaKVPath(t *testing.T) {
	var (
		addr    = "127.0.0.1:50500"
		ch      = make(telemetry.ExtDSChan, 1)
		ctx     = context.Background()
		sensors []*config.Sensor
	)
	ln, err := mock.StartGNMIServer(addr, mock.Update{Notification: mock.AristaUpdate(), Attempt: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	cfg := config.NewMockConfig()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	sensors = append(sensors, &config.Sensor{
		Service: "arista.gnmi",
		Output:  "console::stdout",
		Path:    "/interfaces/interface[name=Ethernet1]/state/counters",
	})

	g := New(cfg.Logger(), conn, sensors, ch)
	g.Start(ctx)

	resp := <-ch

	assert.Equal(t, sensors[0].Path, resp.DS["prefix"].(string))
	assert.Equal(t, "127.0.0.1", resp.DS["system_id"].(string))
	assert.Equal(t, int64(1595363593437180059), resp.DS["timestamp"].(int64))
	assert.Equal(t, "Ethernet1", resp.DS["labels"].(map[string]string)["name"])
	assert.Equal(t, "out-octets", resp.DS["key"].(string))
	assert.Equal(t, int64(50302030597), resp.DS["value"].(int64))
	assert.Equal(t, "console::stdout", resp.Output)

	assert.Equal(t, cfg.LogOutput.String(), "", "unexpected logging")
}

func BenchmarkDS(b *testing.B) {
	var (
		cfg     = config.NewMockConfig()
		buf     = &bytes.Buffer{}
		metrics = make(map[string]status.Metrics)
	)

	metrics["dropsTotal"] = status.NewCounter("arista_gnmi_drops_total", "")

	g := GNMI{
		logger:  cfg.Logger(),
		outChan: make(telemetry.ExtDSChan, 100),
		metrics: metrics,
	}

	g.pathOutput = map[string]string{"/interfaces/interface/state/counters/": "console::stdout"}

	n := mock.AristaUpdate()

	for i := 0; i < b.N; i++ {
		g.datastore(buf, n, n.Update[0], "127.0.0.1")
		<-g.outChan
	}
}

func TestGetValue(t *testing.T) {
	i := &gnmi.TypedValue{Value: &gnmi.TypedValue_IntVal{IntVal: 505}}
	v, err := getValue(i)
	assert.NoError(t, err)
	assert.Equal(t, int64(505), v)

	i = &gnmi.TypedValue{Value: &gnmi.TypedValue_BoolVal{BoolVal: true}}
	v, err = getValue(i)
	assert.NoError(t, err)
	assert.Equal(t, true, v)

	i = &gnmi.TypedValue{Value: &gnmi.TypedValue_BytesVal{BytesVal: []byte("test")}}
	v, err = getValue(i)
	assert.NoError(t, err)
	assert.Equal(t, []byte("test"), v)

	i = &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "test"}}
	v, err = getValue(i)
	assert.NoError(t, err)
	assert.Equal(t, "test", v)

	i = &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: 5}}
	v, err = getValue(i)
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), v)

	i = &gnmi.TypedValue{Value: &gnmi.TypedValue_FloatVal{FloatVal: 5.5}}
	v, err = getValue(i)
	assert.NoError(t, err)
	assert.Equal(t, float32(5.5), v)

	i = &gnmi.TypedValue{Value: &gnmi.TypedValue_AsciiVal{AsciiVal: "test"}}
	v, err = getValue(i)
	assert.NoError(t, err)
	assert.Equal(t, "test", v)

	i = &gnmi.TypedValue{Value: &gnmi.TypedValue_DecimalVal{DecimalVal: &gnmi.Decimal64{Digits: 5, Precision: 1}}}
	v, err = getValue(i)
	assert.NoError(t, err)
	assert.Equal(t, 0.5, v)

	i = &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonIetfVal{JsonIetfVal: []byte("{\"test\":5}")}}
	v, err = getValue(i)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": float64(5)}, v)

	i = &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: []byte("{\"test\":5}")}}
	v, err = getValue(i)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": float64(5)}, v)

	i = &gnmi.TypedValue{Value: &gnmi.TypedValue_LeaflistVal{LeaflistVal: &gnmi.ScalarArray{Element: []*gnmi.TypedValue{
		{Value: &gnmi.TypedValue_UintVal{UintVal: 5}},
	}}}}
	v, err = getValue(i)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}([]interface{}{uint64(5)}), v)
}

func TestVersion(t *testing.T) {
	assert.Equal(t, gnmiVersion, Version())
}
