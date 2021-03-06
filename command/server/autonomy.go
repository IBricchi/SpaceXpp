/*
   Written by Nicholas Pfaff (nicholas.pfaff19@imperial.ac.uk), 2021 - SpaceX++ EEE/EIE 2nd year group project, Imperial College London
*/

package server

import (
	"container/list"
	"fmt"
)

const tileMapUnknownVal = 1

/*
	Return values: isFullyDiscovered bool, destinationRow int, destinationCol int
	isFullyDiscovered is true if there are no more "unknown" tiles on the map (tile value 1).
*/
func getBestNextDestinationCoordinates(tileMap tileMap) (bool, int, int) {
	// Prefer tile with more unknown neighbors => Can discover all using full discovery traversal mode
	for i := 4; i >= 0; i-- {
		if noSuchFieldExists, row, col := getFieldWithMinUnknownNeighborCount(i, tileMap); !noSuchFieldExists {
			return false, row, col
		}
	}

	// The whole map is already discovered
	return true, -1, -1
}

// Return values: noSuchFieldExists, Row, Col
func getFieldWithMinUnknownNeighborCount(minUnknownNeighborCount int, tileMap tileMap) (bool, int, int) {
	for row := 0; row < tileMap.Rows; row++ {
		for col := 0; col < tileMap.Cols; col++ {
			if tileMap.getTile(row, col) != tileMapUnknownVal {
				continue
			}

			if getUnknownNeighborCount(row, col, tileMap) >= minUnknownNeighborCount {
				return false, row, col
			}
		}
	}

	return true, -1, -1
}

func getUnknownNeighborCount(row int, col int, tileMap tileMap) int {
	unknownNeighborCount := 0

	if tileMap.getTile(row+1, col) == tileMapUnknownVal {
		unknownNeighborCount++
	}
	if tileMap.getTile(row-1, col) == tileMapUnknownVal {
		unknownNeighborCount++
	}
	if tileMap.getTile(row, col+1) == tileMapUnknownVal {
		unknownNeighborCount++
	}
	if tileMap.getTile(row, col-1) == tileMapUnknownVal {
		unknownNeighborCount++
	}

	return unknownNeighborCount
}

type rectangle struct {
	left   int
	right  int
	top    int
	bottom int
}

/*
	Expects no rectangles that are touching the map edges as all map edges should be filled by an obstruction.

	More work is needed for this to produce an exact solution when rectangles are touching.
	This algorithm should not be used for automation until improvements are added.
	However, this algorithm could potentially provide a more elegant solution than the current automation one.
*/
func getUndiscoveredRectangles(tileMap tileMap) []rectangle {
	finishedRectangles := make([]rectangle, 0)
	unfinishedRectangles := list.New() // linked list
	var currentListEntry *list.Element
	var currentRectangle *rectangle
	isNewRectangle := false

	for row := 0; row < tileMap.Rows; row++ {
		currentListEntry = unfinishedRectangles.Front()
		if currentListEntry != nil {
			currentRectangle = currentListEntry.Value.(*rectangle)
		}

		for col := 0; col < tileMap.Cols; col++ {
			if tileMap.getTile(row, col) == tileMapUnknownVal { // Part of some rectangle
				if currentListEntry != nil && col == currentRectangle.left { // Start col of current unfinished rectangle
					finishedRectangle := false

					for col <= currentRectangle.right {
						// Current rectangle finished
						if tileMap.getTile(row, col) != tileMapUnknownVal {
							// Reset col for finding new rectangles in this row
							col = currentRectangle.left

							currentRectangle.bottom = row - 1

							// Add current rectangle to finished
							finishedRectangles = append(finishedRectangles, *currentRectangle)

							fmt.Println(currentRectangle)

							// Remove current rectangle from unfinished
							nextListEntry := currentListEntry.Next()
							unfinishedRectangles.Remove(currentListEntry)
							currentListEntry = nextListEntry
							if currentListEntry != nil {
								currentRectangle = currentListEntry.Value.(*rectangle)
							}

							finishedRectangle = true

							break
						}

						col++
					}
					col--

					if !finishedRectangle {
						// Move to next unfinished rectangle
						currentListEntry = currentListEntry.Next()
						if currentListEntry != nil {
							currentRectangle = currentListEntry.Value.(*rectangle)
						}
					}
				} else if !isNewRectangle { // Found a new rectangle
					isNewRectangle = true

					newRectangle := rectangle{
						left:   col,
						right:  col,
						top:    row,
						bottom: row,
					}

					if currentListEntry != nil {
						currentListEntry = unfinishedRectangles.InsertBefore(&newRectangle, currentListEntry)
					} else { // Linked list currently empty
						currentListEntry = unfinishedRectangles.PushBack(&newRectangle)
					}

					currentRectangle = currentListEntry.Value.(*rectangle)
				}
			} else { // Not part of a rectangle
				if isNewRectangle { // Update right border of new rectangle
					isNewRectangle = false

					currentRectangle.right = col - 1

					// Move to next unfinished rectangle
					currentListEntry = currentListEntry.Next()
					if currentListEntry != nil {
						currentRectangle = currentListEntry.Value.(*rectangle)
					}
				} else if currentListEntry != nil && col == currentRectangle.left { // Found bottom of unfinished rectangle => finished
					currentRectangle.bottom = row - 1

					// Add current rectangle to finished
					finishedRectangles = append(finishedRectangles, *currentRectangle)

					// Remove current rectangle from unfinished
					nextListEntry := currentListEntry.Next()
					unfinishedRectangles.Remove(currentListEntry)
					currentListEntry = nextListEntry
					if currentListEntry != nil {
						currentRectangle = currentListEntry.Value.(*rectangle)
					}
				}
			}
		}
	}
	return finishedRectangles
}

// Return values: CentreRow, CentreCol
func getRectangleCentreCoordinates(rectangle rectangle) (int, int) {
	centreRow := (rectangle.top + rectangle.bottom) / 2
	centreCol := (rectangle.left + rectangle.right) / 2

	return centreRow, centreCol
}
