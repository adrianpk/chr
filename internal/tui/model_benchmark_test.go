package tui

import (
	"fmt"
	"testing"
	"unsafe"

	tea "github.com/charmbracelet/bubbletea"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

// BenchmarkModelUpdate measures the event-loop interface with a populated board.
func BenchmarkModelUpdate(b *testing.B) {
	cases := []struct {
		name string
		msg  tea.Msg
	}{
		{name: "idle", msg: struct{}{}},
		{name: "focus", msg: tea.KeyMsg{Type: tea.KeyTab}},
		{name: "resize", msg: tea.WindowSizeMsg{Width: 120, Height: 36}},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			var model tea.Model = benchmarkBoardModel()

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				model, _ = model.Update(tc.msg)
			}

			b.ReportMetric(float64(unsafe.Sizeof(Model{})), "model-bytes")
		})
	}
}

// BenchmarkModelView measures the existing terminal rendering paths at fixed sizes.
func BenchmarkModelView(b *testing.B) {
	for _, sized := range []bool{false, true} {
		b.Run(fmt.Sprintf("shell=%t", sized), func(b *testing.B) {
			model := benchmarkBoardModel()
			if sized {
				model.width, model.height = 120, 36
			}

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				_ = model.View()
			}
		})
	}
}

func benchmarkBoardModel() *Model {
	tickets := make([]ticketcore.Indexed, 128)
	for i := range tickets {
		tickets[i].Key = fmt.Sprintf("TKT-%04d", i)
		tickets[i].Title = fmt.Sprintf("Ticket %d", i)
		tickets[i].Status = "ready"
		tickets[i].Priority = "normal"
	}

	model := NewModel(Options{ProjectDir: "/project", Tickets: tickets})
	model.selectView(ViewBoard)

	return model
}
