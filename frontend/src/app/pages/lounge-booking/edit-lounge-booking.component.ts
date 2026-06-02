import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { LoungeBookingService } from '../../core/services/lounge-booking.service';
import { LoungeBooking } from '../../core/models/lounge-booking.model';

@Component({
  selector: 'app-edit-lounge-booking',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './edit-lounge-booking.component.html',
  styleUrls: ['./edit-lounge-booking.component.scss']
})
export class EditLoungeBookingComponent implements OnInit {
  booking?: LoungeBooking;

  // Editable fields (not all fields are editable after backend change)
  formGuests = 0;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private svc: LoungeBookingService
  ) {}

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (!id) { this.router.navigate(['/lounge-booking']); return; }
    const found = this.svc.bookings.find(b => b.lounge_booking_id === id);
    if (!found) { this.router.navigate(['/lounge-booking']); return; }
    this.booking = { ...found };

    // Initialize form from existing booking
    this.formGuests = this.booking.number_of_guests;
  }

  save(): void {
    if (!this.booking) return;

    // Only allow changes to number_of_guests
    const updated: LoungeBooking = {
      ...this.booking,
      number_of_guests: Math.max(0, Number(this.formGuests) || 0),
    };

    this.svc.update(updated);
    this.router.navigate(['/lounge-booking']);
  }

  cancel(): void {
    this.router.navigate(['/lounge-booking']);
  }

  // Helpers
  get startDateTime(): string {
    if (!this.booking) return '';
    // Display as local datetime string for readability
    const d = new Date(this.booking.scheduled_arrival);
    return d.toLocaleString();
  }
}