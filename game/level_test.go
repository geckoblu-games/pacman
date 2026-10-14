package game

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// TestLevelParametersMatchDossier compares every parameter row with the local source table.
func TestLevelParametersMatchDossier(t *testing.T) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test source")
	}
	dossier, err := os.ReadFile(filepath.Join(filepath.Dir(sourceFile), "..", "doc", "The Pac-Man Dossier.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, table, found := strings.Cut(string(dossier), "### Table A.1 \u2014 Level Specifications")
	if !found {
		t.Fatal("missing table A.1")
	}
	table, _, _ = strings.Cut(table, "### Table A.2")
	parseValue := func(value string) int {
		t.Helper()
		if value == "\u2013" {
			return 0
		}
		number, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(value, "~"), "%"))
		if err != nil {
			t.Fatalf("invalid table value %q: %v", value, err)
		}
		return number
	}
	rows := 0
	for _, line := range strings.Split(table, "\n") {
		fields := strings.Split(line, "|")
		if len(fields) != 18 {
			continue
		}
		fields = fields[1 : len(fields)-1]
		for index := range fields {
			fields[index] = strings.TrimSpace(fields[index])
		}
		number, err := strconv.Atoi(strings.TrimSuffix(fields[0], "+"))
		if err != nil {
			continue
		}
		rows++
		if number != rows {
			t.Fatalf("unexpected table row %d: level %d", rows, number)
		}
		want := LevelParameters{
			BonusSymbol:                  fields[1],
			BonusPoints:                  parseValue(fields[2]),
			PacManSpeedPercent:           parseValue(fields[3]),
			PacManDotsSpeedPercent:       parseValue(fields[4]),
			GhostSpeedPercent:            parseValue(fields[5]),
			GhostTunnelSpeedPercent:      parseValue(fields[6]),
			Elroy1DotsLeft:               parseValue(fields[7]),
			Elroy1SpeedPercent:           parseValue(fields[8]),
			Elroy2DotsLeft:               parseValue(fields[9]),
			Elroy2SpeedPercent:           parseValue(fields[10]),
			FrightPacManSpeedPercent:     parseValue(fields[11]),
			FrightPacManDotsSpeedPercent: parseValue(fields[12]),
			FrightGhostSpeedPercent:      parseValue(fields[13]),
			FrightTimeSeconds:            parseValue(fields[14]),
			FrightFlashes:                parseValue(fields[15]),
		}
		if got := NewLevelState(number).Parameters; got != want {
			t.Errorf("level %d parameters: got %+v, want %+v", number, got, want)
		}
	}
	if rows != 21 {
		t.Fatalf("expected 21 dossier rows, got %d", rows)
	}
	if len(levelParameters) != rows {
		t.Fatalf("expected 21 table rows, got dossier=%d code=%d", rows, len(levelParameters))
	}
	if BaseSpeedPixelsPerSecond != 75.75757625 {
		t.Fatal("base speed differs from table A.1")
	}
}

// TestCollectibles verifies initial counts, invalid positions, points, and idempotence.
func TestCollectibles(t *testing.T) {
	state := NewLevelState(1)
	if state.Remaining != 244 || len(state.Coins) != 240 || len(state.Energizers) != 4 {
		t.Fatal("expected 240 coins and four energizers")
	}
	if state.ConsumeCollectible(-1, -1) != 0 || state.Remaining != 244 {
		t.Fatal("invalid positions should have no effect")
	}
	for index, coord := range CoinCoords {
		if !state.Coins[index] || state.ConsumeCollectible(coord[0], coord[1]) != 10 ||
			state.Coins[index] || state.ConsumeCollectible(coord[0], coord[1]) != 0 {
			t.Fatal("each coin should award ten points exactly once")
		}
	}
	if state.Remaining != 4 {
		t.Fatal("energizers are required to finish the level")
	}
	for index, coord := range EnergizerCoords {
		if !state.Energizers[index] || state.ConsumeCollectible(coord[0], coord[1]) != 50 ||
			state.Energizers[index] || state.ConsumeCollectible(coord[0], coord[1]) != 0 {
			t.Fatal("each energizer should award fifty points exactly once")
		}
	}
	if state.Remaining != 0 {
		t.Fatal("all collectibles should be consumed")
	}
}
