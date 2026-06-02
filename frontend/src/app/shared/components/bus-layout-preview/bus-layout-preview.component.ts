import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

interface SeatInfo {
  position: number;
  seat_number: string;
  is_window_seat: boolean;
  is_aisle_seat: boolean;
}

interface BusRow {
  row_number: number;
  row_label: string;
  left_seats: SeatInfo[];
  right_seats: SeatInfo[];
}

interface BusLayoutPreview {
  rows: BusRow[];
}

@Component({
  selector: 'app-bus-layout-preview',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './bus-layout-preview.component.html',
  styleUrls: ['./bus-layout-preview.component.scss']
})
export class BusLayoutPreviewComponent {
  @Input() layoutPreview!: BusLayoutPreview;
}
