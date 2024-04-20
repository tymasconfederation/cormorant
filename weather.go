package cormorant

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/innotechdevops/openmeteo"
	"github.com/tymasconfederation/cormorant/pb"
	"google.golang.org/protobuf/proto"
)

// Geocode calls the open-meteo geocoding API to get information a postal code or place name,
// and returns a struct containing that information unless an error occurred or we failed to find anything.
func Geocode(place string) (ret *pb.GeocodingApi_Geoname, err error) {
	uri := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%v&count=1&language=en&format=protobuf", url.QueryEscape(place))
	var resp []byte
	if resp, err = call(uri); err == nil {
		var msg *pb.GeocodingApi_SearchResults = &pb.GeocodingApi_SearchResults{}
		if err = proto.Unmarshal(resp, msg); err == nil {
			results := msg.GetResults()
			if len(results) == 0 {
				err = fmt.Errorf("Unable to find `%v`", place)
			} else if len(results) >= 1 {
				ret = results[0]
			}
		}
	}
	return
}

// Forecast returns a forecast for a location, or an error if something goes wrong.
func Forecast(place string) (ret string, err error) {
	var geo *pb.GeocodingApi_Geoname
	if geo, err = Geocode(place); err == nil {
		if geo == nil {
			err = fmt.Errorf("Geocode(%v) returned nil for both geo and err.", place)
		} else {
			param := openmeteo.Parameter{
				Latitude:  openmeteo.Float32(geo.Latitude),
				Longitude: openmeteo.Float32(geo.Longitude),
				Elevation: openmeteo.Float32(geo.Elevation),
				Timezone:  openmeteo.String(geo.Timezone),
				Daily: &[]string{
					openmeteo.DailyTemperature2mMin,
					openmeteo.DailyTemperature2mMax,
					openmeteo.DailyApparentTemperatureMin,
					openmeteo.DailyApparentTemperatureMax,
					openmeteo.DailyRainSum,
					openmeteo.DailyShowersSum,
					openmeteo.DailySnowfallSum,
					openmeteo.DailyPrecipitationProbabilityMean,
					openmeteo.DailySunrise,
					openmeteo.DailySunset,
				},
				CurrentWeather: openmeteo.Bool(true),
			}
			m := openmeteo.New()
			var resp string
			if resp, err = m.Execute(param); err == nil {
				var respMap map[string]interface{} = nil
				if err = json.Unmarshal([]byte(resp), &respMap); err == nil {
					curWeather := respMap["current_weather"].(map[string]interface{})
					curTemp := curWeather["temperature"].(float64)
					curWindspeed := curWeather["windspeed"].(float64)
					// curWindDir := curWeather["winddirection"].(float64)
					weatherCode := curWeather["weathercode"].(float64)

					dailyWeather := respMap["daily"].(map[string]interface{})
					minTemp := dailyWeather["temperature_2m_min"].([]interface{})[0].(float64)
					maxTemp := dailyWeather["temperature_2m_max"].([]interface{})[0].(float64)
					minTempApparent := dailyWeather["apparent_temperature_min"].([]interface{})[0].(float64)
					maxTempApparent := dailyWeather["apparent_temperature_max"].([]interface{})[0].(float64)
					rainSum := dailyWeather["rain_sum"].([]interface{})[0].(float64)
					showersSum := dailyWeather["showers_sum"].([]interface{})[0].(float64)
					snowfallSum := dailyWeather["snowfall_sum"].([]interface{})[0].(float64)
					precipChance := dailyWeather["precipitation_probability_mean"].([]interface{})[0].(float64)
					sunrise := dailyWeather["sunrise"].([]interface{})[0].(string)
					sunset := dailyWeather["sunset"].([]interface{})[0].(string)
					idx := strings.IndexRune(sunrise, 'T')
					if idx > -1 {
						sunrise = sunrise[idx+1:]
					}
					idx = strings.IndexRune(sunset, 'T')
					if idx > -1 {
						sunset = sunset[idx+1:]
					}
					curTempF := Fahrenheit(curTemp)
					minTempF := Fahrenheit(minTemp)
					maxTempF := Fahrenheit(maxTemp)
					minTempApparentF := Fahrenheit(minTempApparent)
					maxTempApparentF := Fahrenheit(maxTempApparent)
					weatherCodeStr := openmeteo.WeatherCodeName(int(weatherCode))
					curWindspeedMph := Mph(curWindspeed)
					precipStr := ""
					if precipChance > 0.1 {
						precipStr = fmt.Sprintf("%0.2f%% chance of precipitation", precipChance)
						rainSumI := Inches(rainSum)
						showersSumI := Inches(showersSum)
						snowfallSumI := Inches(snowfallSum)
						if rainSum > 0 {
							precipStr = fmt.Sprintf("%v | %0.2f inches / %0.2f mm rain", precipStr, rainSumI, rainSum)
						}
						if showersSum > 0 {
							precipStr = fmt.Sprintf("%v | %0.2f inches / %0.2f mm showers", precipStr, showersSumI, showersSum)
						}
						if snowfallSum > 0 {
							precipStr = fmt.Sprintf("%v | %0.2f inches / %0.2f mm snowfall", precipStr, snowfallSumI, snowfallSum)
						}
						precipStr = fmt.Sprintf("%v\n", precipStr)
					}
					ret = fmt.Sprintf("Weather for %v, %v, %v: %v\n"+
						"Temperature: Currently %0.2f°F / %0.2f°C, low of %0.2f°F / %0.2f°C (Apparent %0.2f°F / %0.2f°C), high of %0.2f°F / %0.2f°C (Apparent %0.2f°F / %0.2f°C)\n"+
						"%v"+
						"Wind: %0.2f km/h / %0.2f MPH\n"+
						"Sunrise at %v, sunset at %v",
						geo.Name, geo.Admin1, geo.Country, weatherCodeStr, curTempF, curTemp, minTempF, minTemp, maxTempF, maxTemp,
						minTempApparentF, minTempApparent, maxTempApparentF, maxTempApparent, precipStr,
						curWindspeed, curWindspeedMph, sunrise, sunset)
				}
			}
		}
	}
	return
}

// call uses HTTP get to call into an API endpoint and returns the body of the response unless an error occurred.
func call(uri string) (ret []byte, err error) {
	var resp *http.Response
	if resp, err = http.Get(uri); err == nil {
		defer resp.Body.Close()
		ret, err = io.ReadAll(resp.Body)
	}
	return
}

// Fahrenheit converts celsius to fahrenheit
func Fahrenheit(celsius float64) float64 {
	return celsius*1.8 + 32.0
}

// Inches converts millimeters to inches
func Inches(mm float64) float64 {
	return mm / 25.4
}

// Mph converts km/h to MPH
func Mph(kmPerH float64) float64 {
	return kmPerH / 1.609344
}
