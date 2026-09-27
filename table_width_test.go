package main

import "testing"

func TestComputeAssignedWidthsReservesUsageBarSpace(t *testing.T) {
	oldWidth := *width
	defer func() { *width = oldWidth }()

	*width = 120

	opts := TableOptions{
		Columns: []int{1, 2, 3, 4, 5, 10, 11},
	}

	// Simulate a filesystem with an unusually long mount point and
	// filesystem description. Flexible columns should wrap rather than
	// consume the space needed for the usage bar.
	maxColContent := map[int]int{
		1:  160, // Mounted on
		2:  7,   // Size
		3:  7,   // Used
		4:  7,   // Avail
		5:  6,   // Use%
		10: 8,   // Type
		11: 120, // Filesystem
	}

	assigned, slack := computeAssignedWidths(maxColContent, opts)

	wantSlack := maxUsageBarWidth + 1
	if slack < wantSlack {
		t.Fatalf(
			"usage bar space not preserved: got slack=%d, want at least %d",
			slack,
			wantSlack,
		)
	}

	if assigned[1] >= maxColContent[1] {
		t.Fatalf(
			"mount point was not constrained: assigned=%d content=%d",
			assigned[1],
			maxColContent[1],
		)
	}

	if assigned[11] >= maxColContent[11] {
		t.Fatalf(
			"filesystem was not constrained: assigned=%d content=%d",
			assigned[11],
			maxColContent[11],
		)
	}
}
