package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type WeatherData struct {
	Name string `json:"name"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Sys struct {
		Country string `json:"country"`
		Sunrise int64  `json:"sunrise"`
		Sunset  int64  `json:"sunset"`
	} `json:"sys"`
}

type ApiConfigData struct {
	OpenWeatherApiKey string `json:"OpenWeatherApiKey"`
}

func loadApiConfig(filename string) (ApiConfigData, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return ApiConfigData{}, err
	}
	var api ApiConfigData
	err = json.Unmarshal(bytes, &api)
	if err != nil {
		return ApiConfigData{}, err
	}
	return api, nil
}

func main() {
	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		w.Write([]byte("Welcome to the homepage, navigate to /weather/%your-query%"))
	})
	router.HandleFunc("/weather/{city}", func(w http.ResponseWriter, r *http.Request) {
		city := r.PathValue("city")
		data, err := query(city)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write( []byte(data.FormatOutput()))
	})
	s := &http.Server{
		Addr:    ":8070",
		Handler: router,
	}
	fmt.Println("Server Running on http://localhost:8070")
	log.Fatal(s.ListenAndServe())
}

func query(city string) (WeatherData, error) {
	apiConfig, err := loadApiConfig(".apiConfig")
	if err != nil {
		return WeatherData{}, err
	}
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?APPID=%s&q=%s", apiConfig.OpenWeatherApiKey, city)
	resp, err := http.Get(url)
	if err != nil {
		return WeatherData{}, err
	}
	defer resp.Body.Close()

	// Read the entire response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return WeatherData{}, err
	}

	var weather WeatherData
	if err := json.Unmarshal(body, &weather); err != nil {
		return WeatherData{}, err
	}

	return weather, nil
}

func getWeatherEmoji(condition string) string {
	switch strings.ToLower(condition) {
	case "clear":
		return "☀️"
	case "clouds":
		return "☁️"
	case "rain":
		return "🌧️"
	case "drizzle":
		return "🌦️"
	case "thunderstorm":
		return "⛈️"
	case "snow":
		return "❄️"
	case "mist", "fog":
		return "🌫️"
	default:
		return "🌈"
	}
}

func (w WeatherData) FormatOutput() string {
	var output strings.Builder

	fmt.Fprintf(&output, "Weather Report for %s, %s 🌍\n", w.Name, w.Sys.Country)
	fmt.Fprintf(&output, "==================================\n")
	fmt.Fprintf(&output, "Temperature: %.2f°C (%.2f°F) 🌡️\n", w.Main.Temp-273.15, (w.Main.Temp-273.15)*9/5+32)
	fmt.Fprintf(&output, "Feels like: %.2f°C (%.2f°F) 🤔\n", w.Main.FeelsLike-273.15, (w.Main.FeelsLike-273.15)*9/5+32)
	fmt.Fprintf(&output, "Min/Max: %.2f°C / %.2f°C 📊\n", w.Main.TempMin-273.15, w.Main.TempMax-273.15)
	fmt.Fprintf(&output, "Humidity: %d%% 💧\n", w.Main.Humidity)
	fmt.Fprintf(&output, "Pressure: %d hPa 🔬\n", w.Main.Pressure)

	if len(w.Weather) > 0 {
		emoji := getWeatherEmoji(w.Weather[0].Main)
		fmt.Fprintf(&output, "Condition: %s %s (%s)\n", emoji, w.Weather[0].Main, w.Weather[0].Description)
	}

	fmt.Fprintf(&output, "Wind: %.1f m/s, Direction: %d° 🌬️\n", w.Wind.Speed, w.Wind.Deg)
	fmt.Fprintf(&output, "Cloudiness: %d%% ☁️\n", w.Clouds.All)

	sunrise := time.Unix(w.Sys.Sunrise, 0).Format("15:04")
	sunset := time.Unix(w.Sys.Sunset, 0).Format("15:04")
	fmt.Fprintf(&output, "Sunrise: %s 🌅, Sunset: %s 🌇\n", sunrise, sunset)

	return output.String()
}
