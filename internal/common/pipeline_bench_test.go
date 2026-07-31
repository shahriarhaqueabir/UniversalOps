package common

import (
	"testing"
)

func BenchmarkPipelinePush(b *testing.B) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 1000})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dp.PushMetric("cpu.percent", "%", float64(i%100))
	}
}

func BenchmarkPipelineGetTimeSeries(b *testing.B) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 1000})
	for i := 0; i < 500; i++ {
		dp.PushMetric("cpu.percent", "%", float64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dp.GetTimeSeries("cpu.percent")
	}
}

func BenchmarkPipelineGetForecast(b *testing.B) {
	dp := NewDataPipeline(CollectionConfig{
		Capacity:       100,
		ForecastSteps:  5,
		ForecastWindow: 20,
	})
	for i := 0; i < 50; i++ {
		dp.PushMetric("cpu.percent", "%", float64(50+i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dp.GetForecast("cpu.percent", 5)
	}
}

func BenchmarkPipelineGetTrend(b *testing.B) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 100, ForecastWindow: 20})
	for i := 0; i < 20; i++ {
		dp.PushMetric("memory.percent", "%", float64(90-i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dp.GetTrend("memory.percent")
	}
}

func BenchmarkPipelineNumSeries(b *testing.B) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 100})
	dp.PushMetric("cpu.percent", "%", 42.0)
	dp.PushMetric("memory.percent", "%", 65.0)
	dp.PushMetric("disk.percent", "%", 80.0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dp.NumSeries()
	}
}
