import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { DriverService } from '../../core/services/driver.service';
import { Driver } from '../../core/models/driver.model';

interface AddDriverForm {
  name: string;
  contact_number: string;
  license_number: string;
  license_expiry_date: string;
  experience_years: number | null;
  status: 'Active' | 'Inactive' | 'On Leave';
  hire_date: string;
}

interface ValidationErrors {
  name?: string | null;
  contact_number?: string | null;
  license_number?: string | null;
  license_expiry_date?: string | null;
  experience_years?: string | null;
  hire_date?: string | null;
}

@Component({
  selector: 'app-add-driver',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './add-driver.component.html',
  styleUrls: ['./add-driver.component.scss']
})
export class AddDriverComponent {
  form: AddDriverForm = {
    name: '',
    contact_number: '',
    license_number: '',
    license_expiry_date: '',
    experience_years: null,
    status: 'Active',
    hire_date: ''
  };

  isSubmitting = false;
  isFormSubmitted = false;
  validationErrors: ValidationErrors = {};

  constructor(private router: Router, private driverService: DriverService) {}

  // Validation methods
  private validateName(name: string): string | null {
    if (!name || name.trim() === '') return 'Name is required';
    if (name.trim().length < 2) return 'Name must be at least 2 characters';
    const pattern = /^[A-Za-z\s'-]{2,50}$/;
    if (!pattern.test(name.trim())) return 'Name can only contain letters, spaces, hyphens and apostrophes';
    return null;
  }

  private validateContactNumber(phone: string): string | null {
    if (!phone || phone.trim() === '') return 'Contact Number is required';
    const pattern = /^[+]?\d[\d\s-]{6,14}\d$/; // 8-16 digits incl. separators
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
    // must be strictly in the future
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
    // cannot be in the future
    if (d > new Date(today.getFullYear(), today.getMonth(), today.getDate())) return 'Hire Date cannot be in the future';
    return null;
  }

  // Real-time handlers
  onNameChange(): void { this.validationErrors.name = this.validateName(this.form.name); }
  onContactNumberChange(): void { this.validationErrors.contact_number = this.validateContactNumber(this.form.contact_number); }
  onLicenseNumberChange(): void { this.validationErrors.license_number = this.validateLicenseNumber(this.form.license_number); }
  onLicenseExpiryDateChange(): void { this.validationErrors.license_expiry_date = this.validateLicenseExpiryDate(this.form.license_expiry_date); }
  onExperienceChange(): void { this.validationErrors.experience_years = this.validateExperience(this.form.experience_years); }
  onHireDateChange(): void { this.validationErrors.hire_date = this.validateHireDate(this.form.hire_date); }

  isFormValid(): boolean {
    return !this.validateName(this.form.name) &&
           !this.validateContactNumber(this.form.contact_number) &&
           !this.validateLicenseNumber(this.form.license_number) &&
           !this.validateLicenseExpiryDate(this.form.license_expiry_date) &&
           !this.validateExperience(this.form.experience_years) &&
           !this.validateHireDate(this.form.hire_date);
  }

  save(): void {
    this.isFormSubmitted = true;

    // validate all
    this.validationErrors = {
      name: this.validateName(this.form.name),
      contact_number: this.validateContactNumber(this.form.contact_number),
      license_number: this.validateLicenseNumber(this.form.license_number),
      license_expiry_date: this.validateLicenseExpiryDate(this.form.license_expiry_date),
      experience_years: this.validateExperience(this.form.experience_years),
      hire_date: this.validateHireDate(this.form.hire_date)
    };

    if (!this.isFormValid()) {
      return;
    }

    this.isSubmitting = true;
    const newDriver: Driver = {
      id: '', // Backend will generate
      name: this.form.name.trim(),
      contact_number: this.form.contact_number.trim(),
      license_number: this.form.license_number.trim(),
      license_expiry_date: this.form.license_expiry_date,
      experience_years: this.form.experience_years as number,
      verification_status: 'Pending',
      verification_notes: '',
      status: this.form.status,
      hire_date: this.form.hire_date
    };
    this.driverService.addDriver(newDriver);
    this.isSubmitting = false;
    this.router.navigate(['/driver-management']);
  }

  cancel(): void {
    this.router.navigate(['/driver-management']);
  }

  goBack(): void {
    this.router.navigate(['/driver-management']);
  }
}


