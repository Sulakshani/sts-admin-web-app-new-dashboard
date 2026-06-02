import { Component, Input, Output, EventEmitter, OnInit, OnChanges, SimpleChanges } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-seat-selector',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './seat-selector.component.html',
  styleUrls: ['./seat-selector.component.scss']
})
export class SeatSelectorComponent implements OnInit, OnChanges {
  @Input() totalRows: number = 5;
  @Output() seatMapChange = new EventEmitter<boolean[][]>();
  @Output() totalRowsChange = new EventEmitter<number>();

  seatMap: boolean[][] = [];
  seatsPerRow = 6; // 3 left + 3 right

  ngOnInit() {
    this.initializeSeatMap();
  }

  ngOnChanges(changes: SimpleChanges) {
    if (changes['totalRows']) {
      this.adjustSeatMapForRowChange();
    }
  }

  initializeSeatMap() {
    // Initialize with default pattern: window seats selected (X X _ _ X X)
    this.seatMap = Array(this.totalRows).fill(null).map(() =>
      [true, true, false, false, true, true]
    );
    this.emitSeatMap();
  }

  adjustSeatMapForRowChange() {
    const currentRows = this.seatMap.length;
    const newRows = this.totalRows;

    if (newRows > currentRows) {
      // Add new rows at the end with default pattern (X X _ _ X X)
      for (let i = currentRows; i < newRows; i++) {
        this.seatMap.push([true, true, false, false, true, true]);
      }
    } else if (newRows < currentRows) {
      // Remove rows from the end
      this.seatMap = this.seatMap.slice(0, newRows);
    }

    this.emitSeatMap();
  }

  toggleSeat(rowIndex: number, seatIndex: number) {
    this.seatMap[rowIndex][seatIndex] = !this.seatMap[rowIndex][seatIndex];
    this.emitSeatMap();
  }

  emitSeatMap() {
    this.seatMapChange.emit([...this.seatMap.map(row => [...row])]);
  }

  getRowLabel(rowIndex: number): string {
    // Convert 0-based index to alphabetic label (A, B, C...)
    let label: string;
    if (rowIndex < 26) {
      label = String.fromCharCode(65 + rowIndex);
    } else {
      // For rows > 26, use AA, AB, etc.
      const firstLetter = String.fromCharCode(65 + Math.floor(rowIndex / 26) - 1);
      const secondLetter = String.fromCharCode(65 + (rowIndex % 26));
      label = firstLetter + secondLetter;
    }

    // Add row number in brackets
    const rowNumber = rowIndex + 1;
    const suffix = this.getOrdinalSuffix(rowNumber);
    return `${label} (${rowNumber}${suffix} row)`;
  }

  getOrdinalSuffix(num: number): string {
    const j = num % 10;
    const k = num % 100;
    if (j === 1 && k !== 11) return 'st';
    if (j === 2 && k !== 12) return 'nd';
    if (j === 3 && k !== 13) return 'rd';
    return 'th';
  }

  isSeatSelected(rowIndex: number, seatIndex: number): boolean {
    return this.seatMap[rowIndex]?.[seatIndex] ?? false;
  }

  getSeatPositionClass(seatIndex: number): string {
    if (seatIndex === 0 || seatIndex === 5) return 'window-seat';
    if (seatIndex === 2 || seatIndex === 3) return 'aisle-seat';
    return 'middle-seat';
  }

  getTotalSeats(): number {
    return this.seatMap.reduce((total, row) =>
      total + row.filter(seat => seat).length, 0
    );
  }

  getActualSeatNumber(rowIndex: number, seatIndex: number): number {
    // Calculate continuous seat number based on selected seats in this row
    let seatNumber = 0;
    for (let i = 0; i <= seatIndex; i++) {
      if (this.seatMap[rowIndex][i]) {
        seatNumber++;
      }
    }
    return seatNumber;
  }

  addRow() {
    this.totalRows++;
    this.seatMap.push([true, true, false, false, true, true]);
    this.totalRowsChange.emit(this.totalRows);
    this.emitSeatMap();
  }

  removeRow() {
    if (this.totalRows > 1) {
      this.totalRows--;
      this.seatMap.pop();
      this.totalRowsChange.emit(this.totalRows);
      this.emitSeatMap();
    }
  }
}
