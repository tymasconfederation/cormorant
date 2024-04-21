package cormorant

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/innotechdevops/openmeteo"
	"github.com/tymasconfederation/cormorant/pb"
	"google.golang.org/protobuf/proto"
)

var weatherCodeName map[int]string = map[int]string{
	0:  ":sun: Clear sky",
	1:  ":white_sun_small_cloud: Mainly clear",
	2:  ":partly_cloudy: Partly cloudy",
	3:  ":cloud: Overcast",
	45: ":fog: Fog",
	48: ":fog: Depositing rime fog",
	51: ":cloud_rain: Light drizzle",
	53: ":cloud_rain: Moderate drizzle",
	55: ":cloud_rain: Dense drizzle",
	56: ":cloud_rain: Light freezing drizzle",
	57: ":cloud_rain: Dense freezing drizzle",
	61: ":cloud_rain: Slight rain",
	63: ":cloud_rain: Moderate rain",
	65: ":cloud_rain: Heavy rain",
	66: ":cloud_rain: Light freezing rain",
	67: ":cloud_rain: Heavy freezing rain",
	71: ":cloud_snow: Slight snowfall",
	73: ":cloud_snow: Moderate snowfall",
	75: ":cloud_snow: Heavy snowfall",
	77: ":cloud_snow: Snow grains",
	80: ":cloud_rain: Slight rain showers",
	81: ":cloud_rain: Moderate rain showers",
	82: ":cloud_rain: Violent rain showers",
	85: ":cloud_snow: Slight snow showers",
	86: ":cloud_snow: Heavy snow showers",
	95: ":thunder_cloud_rain: Thunderstorm",
	96: ":thunder_cloud_rain: Thunderstorm with slight hail",
	99: ":thunder_cloud_rain: Thunderstorm with heavy hail",
}

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
func Forecast(place string, forecast int) (ret string, err error) {
	var geo *pb.GeocodingApi_Geoname
	if geo, err = Geocode(place); err == nil {
		if geo == nil {
			err = fmt.Errorf("Geocode(%v) returned nil for both geo and err.", place)
		} else {
			switch forecast {
			case CurrentForecast:
				param := openmeteo.Parameter{
					Latitude:  openmeteo.Float32(geo.Latitude),
					Longitude: openmeteo.Float32(geo.Longitude),
					Elevation: openmeteo.Float32(geo.Elevation),
					Timezone:  openmeteo.String(geo.Timezone),
					Daily: &[]string{
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
						weatherCodeStr := weatherCodeName[int(weatherCode)]
						curWindspeedMph := Mph(curWindspeed)
						fmt.Printf("Weather code %v, weatherCodeStr %v\n", int(weatherCode), weatherCodeStr)
						ret = fmt.Sprintf("Current conditions for %v, %v, %v:\n%v\n"+
							":thermometer: Temperature: Currently %0.1f°F (%0.1f°C).\n"+
							":dash: Wind: %0.2f MPH / %0.2f km/h\n"+
							":sunrise: Sunrise at %v\n"+
							":city_dusk: Sunset at %v",
							geo.Name, geo.Admin1, geo.Country, weatherCodeStr, curTempF, curTemp,
							curWindspeedMph, curWindspeed, sunrise, sunset)
					}
				}
			case TodayForecast:
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
						weatherCodeStr := weatherCodeName[int(weatherCode)]
						curWindspeedMph := Mph(curWindspeed)
						precipStr := ""
						if precipChance > 0.1 {
							precipStr = fmt.Sprintf(":8ball: %0.0f%% chance of precipitation", precipChance)
							rainSumI := Inches(rainSum)
							showersSumI := Inches(showersSum)
							snowfallSumI := Inches(snowfallSum)
							if rainSum > 0 {
								precipStr = fmt.Sprintf("%v | :umbrella: %0.2f inches / %0.2f mm rain", precipStr, rainSumI, rainSum)
							}
							if showersSum > 0 {
								precipStr = fmt.Sprintf("%v | :droplet: %0.2f inches / %0.2f mm showers", precipStr, showersSumI, showersSum)
							}
							if snowfallSum > 0 {
								precipStr = fmt.Sprintf("%v | :snowflake: %0.2f inches / %0.2f mm snowfall", precipStr, snowfallSumI, snowfallSum)
							}
							precipStr = fmt.Sprintf("%v\n", precipStr)
						}
						ret = fmt.Sprintf("Weather for %v, %v, %v:\n%v\n"+
							":thermometer: Temperature: Currently %0.1f°F (%0.1f°C).\n"+
							":arrow_down: Low of %0.1f°F (%0.1f°C), apparent %0.1f°F (*%0.1f°C).\n"+
							":arrow_up: High of %0.1f°F (%0.1f°C), apparent %0.1f°F (%0.1f°C)\n"+
							"%v"+
							":dash: Wind: %0.2f MPH / %0.2f km/h\n"+
							":sunrise: Sunrise at %v\n"+
							":city_dusk: Sunset at %v",
							geo.Name, geo.Admin1, geo.Country, weatherCodeStr, curTempF, curTemp, minTempF, minTemp, maxTempF, maxTemp,
							minTempApparentF, minTempApparent, maxTempApparentF, maxTempApparent, precipStr,
							curWindspeedMph, curWindspeed, sunrise, sunset)
					}
				}
			case WeekForecast:
				param := openmeteo.Parameter{
					Latitude:  openmeteo.Float32(geo.Latitude),
					Longitude: openmeteo.Float32(geo.Longitude),
					Elevation: openmeteo.Float32(geo.Elevation),
					Timezone:  openmeteo.String(geo.Timezone),
					Daily: &[]string{
						openmeteo.DailyWeatherCode,
						openmeteo.DailyTemperature2mMin,
						openmeteo.DailyTemperature2mMax,
						// openmeteo.DailyApparentTemperatureMin,
						// openmeteo.DailyApparentTemperatureMax,
						openmeteo.DailyRainSum,
						openmeteo.DailyShowersSum,
						openmeteo.DailySnowfallSum,
						openmeteo.DailyPrecipitationProbabilityMean,
						openmeteo.DailySunrise,
						openmeteo.DailySunset,
					},
				}
				m := openmeteo.New()
				var resp string
				if resp, err = m.Execute(param); err == nil {
					var respMap map[string]interface{} = nil
					if err = json.Unmarshal([]byte(resp), &respMap); err == nil {
						dailyWeather := respMap["daily"].(map[string]interface{})
						dates := dailyWeather["time"].([]interface{})
						weatherCode := dailyWeather["weathercode"].([]interface{})
						minTemp := dailyWeather["temperature_2m_min"].([]interface{})
						maxTemp := dailyWeather["temperature_2m_max"].([]interface{})
						// minTempApparent := dailyWeather["apparent_temperature_min"].([]interface{})
						// maxTempApparent := dailyWeather["apparent_temperature_max"].([]interface{})
						rainSum := dailyWeather["rain_sum"].([]interface{})
						showersSum := dailyWeather["showers_sum"].([]interface{})
						snowfallSum := dailyWeather["snowfall_sum"].([]interface{})
						precipChance := dailyWeather["precipitation_probability_mean"].([]interface{})
						sunrise := dailyWeather["sunrise"].([]interface{})
						sunset := dailyWeather["sunset"].([]interface{})
						ret = "" // "Day | Weather | Low | High | % Precip | Total rain | Total showers | Total snow"
						for day := 0; day < len(weatherCode); day++ {
							dateStr := dates[day].(string)
							var dayStr = "Today"
							if day == 1 {
								dayStr = "Tomorrow"
							} else if day > 1 {
								if t, err := time.Parse(time.DateOnly, dateStr); err == nil {
									dayStr = t.Weekday().String()
								} else {
									dayStr = err.Error()
									fmt.Printf("Error calling time.Parse(DateOnly, \"%v\": %v\n", dateStr, err.Error())
								}
							}
							weatherCodeD := weatherCode[day].(float64)
							minTempD := minTemp[day].(float64)
							maxTempD := maxTemp[day].(float64)
							// minTempApparentD := minTempApparent[day].(float64)
							// maxTempApparentD := maxTempApparent[day].(float64)
							rainSumD := rainSum[day].(float64)
							showersSumD := showersSum[day].(float64)
							snowfallSumD := snowfallSum[day].(float64)
							precipChanceD := precipChance[day].(float64)

							sunriseD := sunrise[day].(string)
							sunsetD := sunset[day].(string)
							minTempF := Fahrenheit(minTempD)
							maxTempF := Fahrenheit(maxTempD)
							// minTempApparentF := Fahrenheit(minTempApparentD)
							// maxTempApparentF := Fahrenheit(maxTempApparentD)
							weatherCodeStr := weatherCodeName[int(weatherCodeD)]
							rainSumI, showersSumI, snowfallSumI := 0.0, 0.0, 0.0
							if precipChanceD > 0.1 {
								rainSumI = Inches(rainSumD)
								showersSumI = Inches(showersSumD)
								snowfallSumI = Inches(snowfallSumD)
							}
							idx := strings.IndexRune(sunriseD, 'T')
							if idx > -1 {
								sunriseD = sunriseD[idx+1:]
							}
							idx = strings.IndexRune(sunsetD, 'T')
							if idx > -1 {
								sunsetD = sunsetD[idx+1:]
							}
							// Day | Weather | Low | High | % Precip | Total rain | Total showers | Total snow
							rainStr := ""
							showersStr := ""
							snowStr := ""
							if rainSumD > 0 {
								rainStr = fmt.Sprintf(" | :umbrella: %0.2f in (%0.2f mm) rain", rainSumI, rainSumD)
							}
							if showersSumD > 0 {
								showersStr = fmt.Sprintf(" | :droplet: %0.2f in (%0.2f mm) showers", showersSumI, showersSumD)
							}
							if snowfallSumD > 0 {
								snowStr = fmt.Sprintf(" | :snowflake: %0.2f in (%0.2f mm) snowfall", snowfallSumI, snowfallSumD)
							}
							ret = fmt.Sprintf("%v\n"+
								"%v: %v | :arrow_down: Low %0.1f°F (%0.1f°C) | :arrow_up: High %0.1f°F (%0.1f°C) | :8ball: %0.0f%% chance of precipitation%v%v%v", ret,
								dayStr, weatherCodeStr, minTempF, minTempD, maxTempF, maxTempD,
								precipChanceD, rainStr, showersStr, snowStr) // rainSumI, rainSumD, showersSumI, showersSumD, snowfallSumI, snowfallSumD)
						}

					}
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
