import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, ActivatedRoute } from '@angular/router';
import { PassengerService } from '../../core/services/passenger.service';
import { Passenger } from '../../core/models/passenger.model';

@Component({
  selector: 'app-edit-passenger',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './edit-passenger.component.html',
  styleUrls: ['./edit-passenger.component.scss']
})
export class EditPassengerComponent implements OnInit {
  passenger: Passenger = {
    passenger_id: '',
    name: '',
    phone: '',
    email: '',
    nic: '',
    created_at: ''
  };

  constructor(
    private router: Router,
    private route: ActivatedRoute,
    private passengerService: PassengerService
  ) {}

  ngOnInit(): void {
    const passengerId = this.route.snapshot.paramMap.get('id');
    if (passengerId) {
      const foundPassenger = this.passengerService.getPassengerById(passengerId);
      if (foundPassenger) {
        this.passenger = { ...foundPassenger };
      } else {
        console.error('Passenger not found');
        this.router.navigate(['/passenger-management']);
      }
    }
  }

  onSubmit() {
    if (this.isFormValid()) {
      this.passengerService.updatePassenger(this.passenger);
      console.log('Passenger updated:', this.passenger);
      this.router.navigate(['/passenger-management']);
    }
  }

  goBack() {
    this.router.navigate(['/passenger-management']);
  }

  private isFormValid(): boolean {
    return !!(this.passenger.name && this.passenger.phone && this.passenger.email && this.passenger.nic);
  }
}
