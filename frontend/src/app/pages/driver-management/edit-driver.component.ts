
import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { DriverService } from '../../core/services/driver.service';
import { Driver } from '../../core/models/driver.model';

interface ValidationErrors {
  name?: string | null;
  contact_number?: string | null;
  license_number?: string | null;
  license_expiry_date?: string | null;
  experience_years?: string | null;
  hire_date?: string | null;
}

@Component({
  selector: 'app-edit-driver',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './edit-driver.component.html',
  styleUrls: ['./edit-driver.component.scss']
})
export class EditDriverComponent implements OnInit {
  driver: Driver | undefined;
  isSubmitting = false;
  isFormSubmitted = false;
  validationErrors: ValidationErrors = {};

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private driverService: DriverService
  ) {}

  ngOnInit(): void {
    const driverId = this.route.snapshot.paramMap.get('id');
    if (driverId) {
      this.driver = this.driverService.getById(driverId);
      if (!this.driver) {
        this.router.navigate(['/driver-management']);
      }
    }
  }

  // Validation logic (mirrors add-driver)
  private validateName(name: string): string | null {
    if (!name || name.trim() === '') return 'Name is required';
    if (name.trim().length < 2) return 'Name must be at least 2 characters';
    const pattern = /^[A-Za-z\s'-]{2,50}$/;
    if (!pattern.test(name.trim())) return 'Name can only contain letters, spaces, hyphens and apostrophes';
    return null;
  }

  private validateContactNumber(phone: string): string | null {
    if (!phone || phone.trim() === '') return 'Contact Number is required';
    const pattern = /^[+]?\d[\d\s-]{6,14}\d$/;
    if (!pattern.test(phone.trim())) return 'Invalid contact number';
    return null;
  }

  private validateLicenseNumber(license: string): string | null {
    if (!license || license.trim() === '') return 'License Number is required';
    const pattern = /^[A-Z0-9]{6,15}$/i;
    if (!pattern.test(license.trim())) return 'License Number must be 6-15 letters/numbers';
    return null;
  }

  private validateLicenseExpiryDate(dateStr: string): string | null {
    if (!dateStr) return 'License Expiry Date is required';
    const today = new Date();
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return 'Invalid date';
    if (d <= new Date(today.getFullYear(), today.getMonth(), today.getDate())) return 'License expiry must be in the future';
    return null;
  }

  private validateExperience(exp: number | null): string | null {
    if (exp === null || exp === undefined) return 'Experience is required';
    if (!Number.isInteger(exp) || exp < 0) return 'Experience must be a non-negative integer';
    if (exp > 50) return 'Experience cannot exceed 50 years';
    return null;
  }

  private validateHireDate(dateStr: string): string | null {
    if (!dateStr) return 'Hire Date is required';
    const today = new Date();
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return 'Invalid date';
    if (d > new Date(today.getFullYear(), today.getMonth(), today.getDate())) return 'Hire Date cannot be in the future';
    return null;
  }

  // Real-time handlers
  onNameChange(): void { if (this.driver) this.validationErrors.name = this.validateName(this.driver.name); }
  onContactNumberChange(): void { if (this.driver) this.validationErrors.contact_number = this.validateContactNumber(this.driver.contact_number); }
  onLicenseNumberChange(): void { if (this.driver) this.validationErrors.license_number = this.validateLicenseNumber(this.driver.license_number); }
  onLicenseExpiryDateChange(): void { if (this.driver) this.validationErrors.license_expiry_date = this.validateLicenseExpiryDate(this.driver.license_expiry_date); }
  onExperienceChange(): void { if (this.driver) this.validationErrors.experience_years = this.validateExperience(this.driver.experience_years); }
  onHireDateChange(): void { if (this.driver) this.validationErrors.hire_date = this.validateHireDate(this.driver.hire_date); }

  isFormValid(): boolean {
    if (!this.driver) return false;
    return !this.validateName(this.driver.name) &&
           !this.validateContactNumber(this.driver.contact_number) &&
           !this.validateLicenseNumber(this.driver.license_number) &&
           !this.validateLicenseExpiryDate(this.driver.license_expiry_date) &&
           !this.validateExperience(this.driver.experience_years) &&
           !this.validateHireDate(this.driver.hire_date);
  }

  save(): void {
    if (!this.driver) return;

    this.isFormSubmitted = true;
    this.validationErrors = {
      name: this.validateName(this.driver.name),
      contact_number: this.validateContactNumber(this.driver.contact_number),
      license_number: this.validateLicenseNumber(this.driver.license_number),
      license_expiry_date: this.validateLicenseExpiryDate(this.driver.license_expiry_date),
      experience_years: this.validateExperience(this.driver.experience_years),
      hire_date: this.validateHireDate(this.driver.hire_date)
    };

    if (!this.isFormValid()) {
      alert('update successfully');
      return;
    }

    this.isSubmitting = true;
    this.driver.license_number = this.driver.license_number.trim().toUpperCase();
    this.driverService.updateDriver(this.driver).subscribe({
      next: () => {
        this.isSubmitting = false;
        alert('Successfully updated');
        this.router.navigate(['/driver-management']);
      },
      error: (err) => {
        this.isSubmitting = false;
        console.error('Error updating driver', err);
        alert('Failed to update driver');
      }
    });
  }

  cancel(): void {
    this.router.navigate(['/driver-management']);
  }

  goBack(): void {
    this.router.navigate(['/driver-management']);
  }
}

