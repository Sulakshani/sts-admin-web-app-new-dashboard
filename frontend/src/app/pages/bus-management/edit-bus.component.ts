

import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { BusService } from '../../core/services/bus.service';
import { Bus } from '../../core/models/bus.model';

interface ValidationErrors {
  bus_number?: string | null;
  total_seats?: string | null;
  custom_route_name?: string | null;
}

@Component({
  selector: 'app-edit-bus',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './edit-bus.component.html',
  styleUrls: ['./edit-bus.component.scss']
})
export class EditBusComponent implements OnInit {
  bus: Bus | undefined;
  isSubmitting = false;
  validationErrors: ValidationErrors = {};
  isFormSubmitted = false;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private busService: BusService
  ) {}

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id') as string;
    this.bus = this.busService.getById(id);
    if (!this.bus) {
      alert('Bus not found');
      this.router.navigate(['/bus-management']);
    }
  }

  // Validation methods
  validateBusNumber(busNumber: string): string | null {
    if (!busNumber || busNumber.trim() === '') {
      return 'Bus Number is required';
    }
    
    // Check format: 2 letters followed by 4 digits (e.g., ND1234)
    const busNumberPattern = /^[A-Z]{2}\d{4}$/;
    if (!busNumberPattern.test(busNumber.toUpperCase())) {
      return 'Bus Number must be in format: 2 letters followed by 4 digits (e.g., ND1234)';
    }
    
    return null;
  }

  validateTotalSeats(totalSeats: number | null): string | null {
    if (totalSeats === null || totalSeats === undefined) {
      return 'Total Seats is required';
    }
    
    if (!Number.isInteger(totalSeats) || totalSeats <= 0) {
      return 'Total Seats must be a positive integer';
    }
    
    if (totalSeats > 54) {
      return 'Total Seats cannot exceed 54';
    }
    
    return null;
  }

  validateCustomRouteName(routeName: string): string | null {
    if (!routeName || routeName.trim() === '') {
      return 'Route Name is required';
    }
    
    // Check if it's a string (not just numbers)
    if (!isNaN(Number(routeName))) {
      return 'Route Name must be a string value, not just numbers';
    }
    
    return null;
  }

  // Real-time validation
  onBusNumberChange(): void {
    if (this.bus) {
      this.validationErrors.bus_number = this.validateBusNumber(this.bus.bus_number);
    }
  }

  onTotalSeatsChange(): void {
    if (this.bus) {
      this.validationErrors.total_seats = this.validateTotalSeats(this.bus.total_seats);
    }
  }

  onCustomRouteNameChange(): void {
    if (this.bus) {
      this.validationErrors.custom_route_name = this.validateCustomRouteName(this.bus.custom_route_name);
    }
  }

  // Check if form is valid
  isFormValid(): boolean {
    if (!this.bus) return false;
    return !this.validateBusNumber(this.bus.bus_number) &&
           !this.validateTotalSeats(this.bus.total_seats) &&
           !this.validateCustomRouteName(this.bus.custom_route_name);
  }

  save(): void {
    if (!this.bus) return;
    
    this.isFormSubmitted = true;
    
    // Validate all fields
    this.validationErrors = {
      bus_number: this.validateBusNumber(this.bus.bus_number),
      total_seats: this.validateTotalSeats(this.bus.total_seats),
      custom_route_name: this.validateCustomRouteName(this.bus.custom_route_name)
    };

    if (!this.isFormValid()) {
      alert('Please fill all required fields correctly');
      return;
    }

    this.isSubmitting = true;
    // Convert bus number to uppercase before saving
    this.bus.bus_number = this.bus.bus_number.toUpperCase();
    this.busService.updateBus(this.bus).subscribe({
      next: () => {
        this.isSubmitting = false;
        alert('Successfully updated');
        this.router.navigate(['/bus-management']);
      },
      error: (err) => {
        this.isSubmitting = false;
        console.error('Error updating bus', err);
        alert('Failed to update bus');
      }
    });
  }

  cancel(): void {
    this.router.navigate(['/bus-management']);
  }

  goBack(): void {
    this.router.navigate(['/bus-management']);
  }
}
