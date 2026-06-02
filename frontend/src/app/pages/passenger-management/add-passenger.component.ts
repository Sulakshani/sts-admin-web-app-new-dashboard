import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { PassengerService } from '../../core/services/passenger.service';

@Component({
  selector: 'app-add-passenger',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './add-passenger.component.html',
  styleUrls: ['./add-passenger.component.scss']
})
export class AddPassengerComponent {
  passenger = {
    name: '',
    phone: '',
    email: '',
    nic: ''
  };

  constructor(
    private router: Router,
    private passengerService: PassengerService
  ) {}

  onSubmit() {
    if (this.isFormValid()) {
      const newPassenger = {
        passenger_id: this.generatePassengerId(),
        name: this.passenger.name,
        phone: this.passenger.phone,
        email: this.passenger.email,
        nic: this.passenger.nic,
        created_at: new Date().toISOString()
      };

      this.passengerService.addPassenger(newPassenger);
      console.log('Passenger added:', newPassenger);
      this.router.navigate(['/passenger-management']);
    }
  }

  goBack() {
    this.router.navigate(['/passenger-management']);
  }

  private isFormValid(): boolean {
    return !!(this.passenger.name && this.passenger.phone && this.passenger.email && this.passenger.nic);
  }

  private generatePassengerId(): string {
    return 'P' + Math.random().toString(36).substr(2, 8).toUpperCase();
  }
}
