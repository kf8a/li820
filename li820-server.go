package li820

import (
	"bytes"
	"encoding/xml"
	"github.com/prometheus/client_golang/prometheus"
	serial "github.com/tarm/serial"
	"io"
	"log"
	"math/rand"
	"strings"
	"time"
)

var (
	co2Log = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "trailer_co2_ppm",
		Help: "Current CO2 value measured on the licor.",
	})
)

type LICOR struct {
	port     io.ReadWriteCloser
	portname string
	model    string
	site     string
}

type Datum struct {
	CO2       float64   `xml:"co2" json:"co2"`
	H2O       float64   `xml:"h2o" json:"h2o"`
	TimeStamp time.Time `json:"at"`
	Site      string    `json:"site"`
}

//TestSample provides fake test data
func (licor LICOR) TestSampler(c chan Datum) {
	for {
		datum := Datum{
			TimeStamp: time.Now(),
			Site:      "glbrc",
			CO2:       rand.Float64(),
			H2O:       rand.Float64(),
		}
		c <- datum
		time.Sleep(1 * time.Second)
	}
}

//Sampler provides a function to sample the Licor and return the results in a channel
func (licor LICOR) Sampler(c chan Datum) {
	connection := serial.Config{Name: licor.portname, Baud: 9600}
	port, err := serial.OpenPort(&connection)
	defer port.Close()
	licor.port = port
	if err != nil {
		log.Fatal(err)
	}

	for {
		sample := licor.Sample()
		c <- sample
	}
}

func (licor LICOR) Sample() Datum {
	data := licor.waiting()
	data = strings.Join([]string{data, licor.data()}, "")
	datum := licor.parse(data)
	return datum
}

func (licor LICOR) parse(data string) Datum {
	type result struct {
		XMLName xml.Name `xml:licor.model`
		Datum   Datum    `xml:"data"`
	}

	value := new(result)

	err := xml.Unmarshal([]byte(data), &value)
	if err != nil {
		log.Println("error: %v", err)
		value.Datum.CO2 = -1
		value.Datum.H2O = -1
	}

	value.Datum.TimeStamp = time.Now()
	value.Datum.Site = licor.site

	co2Log.Set(float64(value.Datum.CO2))

	return value.Datum
}

func (licor LICOR) read(sep string) string {
	result := new(bytes.Buffer)

	for !strings.Contains(result.String(), sep) {
		buffer := make([]byte, 1024)
		n, err := licor.port.Read(buffer)
		if err != nil {
			log.Println(err)
		}
		result.Write(buffer[:n])
	}
	return result.String()
}

func (licor LICOR) data() string {
	element := strings.Join([]string{"</", licor.model, ">"}, "")
	data := licor.read(element)
	lastIndex := strings.LastIndex(data, element)
	return data[:lastIndex+len(element)]
}

func (licor LICOR) waiting() string {
	element := strings.Join([]string{"<", licor.model, ">"}, "")
	data := licor.read(element)
	lastIndex := strings.LastIndex(data, element)
	return data[lastIndex:]
}

func NewLicor(model string, site string, portname string) LICOR {
	licor := LICOR{}
	licor.model = model
	licor.site = site
	licor.portname = portname
	return licor
}

func init() {
	prometheus.MustRegister(co2Log)
}
