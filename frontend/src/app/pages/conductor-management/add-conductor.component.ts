import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { ConductorService } from '../../core/services/conductor.service';
import { Conductor } from '../../core/models/conductor.model';

interface AddConductorForm {
  name: string;
  contact_number: string;
  experience_years: number | null;
  license_number: string;
  license_expiry_date: string;
  verification_status: 'Verified' | 'Pending' | 'Rejected';
  verification_notes: string;
  status: 'Active' | 'On Leave' | 'Resigned';
  hire_date: string;
}

interface ValidationErrors {
  name?: string | null;
  contact_number?: string | null;
  experience_years?: string | null;
  license_number?: string | null;
  license_expiry_date?: string | null;
  verification_status?: string | null;
  hire_date?: string | null;
}

@Component({
  selector: 'app-add-conductor',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './add-conductor.component.html',
  styleUrls: ['./add-conductor.component.scss']
})
export class AddConductorComponent {
  form: AddConductorForm = {
    name: '',
    contact_number: '',
    experience_years: null,
    license_number: '',
    license_expiry_date: '',
    verification_status: 'Pending',
    verification_notes: '',
    status: 'Active',
    hire_date: ''
  };

  isSubmitting = false;
  isFormSubmitted = false;
  validationErrors: ValidationErrors = {};

  constructor(private router: Router, private conductorService: ConductorService) {}

  // Validation methods
  private validateName(name: string): string | null {
    if (!name || name.trim() === '') return 'Name is required';
    if (name.trim().length < 2) return 'Name must be at least 2 characters';
    const pattern = /^[A-Za-z\s'-]{2,100}$/;
    if (!pattern.test(name.trim())) return 'Name can only contain letters, spaces, hyphens and apostrophes';
    return null;
  }

  private validateContactNumber(phone: string): string | null {
    if (!phone || phone.trim() === '') return 'Contact Number is required';
    const pattern = /^\d{10}$/; // Exactly 10 digits
    if (!pattern.test(phone.trim())) return 'Contact Number must be exactly 10 digits';
    return null;
  }

  private validateExperience(exp: number | null): string | null {
    if (exp === null || exp === undefined) return 'Experience is required';
    if (!Number.isInteger(exp) || exp < 0) return 'Experience must be a non-negative integer';
    if (exp > 50) return 'Experience cannot exceed 50 years';
    return null;
  }

  private validateLicenseNumber(license: string): string | null {
    if (!license || license.trim() === '') return 'License Number is required';
    if (license.trim().length < 3) return 'License Number must be at least 3 characters';
    return null;
  }

  private validateLicenseExpiryDate(dateStr: string): string | null {
    if (!dateStr) return 'License Expire Date is required';
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return 'Invalid date';
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
  onNameChange(): void { this.validationErrors.name = this.validateName(this.form.name); }
  onContactNumberChange(): void { this.validationErrors.contact_number = this.validateContactNumber(this.form.contact_number); }
  onExperienceChange(): void { this.validationErrors.experience_years = this.validateExperience(this.form.experience_years); }
  onLicenseNumberChange(): void { this.validationErrors.license_number = this.validateLicenseNumber(this.form.license_number); }
  onLicenseExpiryChange(): void { this.validationErrors.license_expiry_date = this.validateLicenseExpiryDate(this.form.license_expiry_date); }
  onHireDateChange(): void { this.validationErrors.hire_date = this.validateHireDate(this.form.hire_date); }

  isFormValid(): boolean {
    return !this.validateName(this.form.name) &&
           !this.validateContactNumber(this.form.contact_number) &&
           !this.validateExperience(this.form.experience_years) &&
           !this.validateLicenseNumber(this.form.license_number) &&
           !this.validateLicenseExpiryDate(this.form.license_expiry_date) &&
           !this.validateHireDate(this.form.hire_date);
  }

  save(): void {
    this.isFormSubmitted = true;

    // Validate all fields
    this.validationErrors = {
      name: this.validateName(this.form.name),
      contact_number: this.validateContactNumber(this.form.contact_number),
      experience_years: this.validateExperience(this.form.experience_years),
      license_number: this.validateLicenseNumber(this.form.license_number),
      license_expiry_date: this.validateLicenseExpiryDate(this.form.license_expiry_date),
      hire_date: this.validateHireDate(this.form.hire_date)
    };

    if (!this.isFormValid()) {
      return;
    }

    this.isSubmitting = true;
    const newConductor: Conductor = {
      id: '', // Backend will generate
      name: this.form.name.trim(),
      contact_number: this.form.contact_number.trim(),
      experience_years: this.form.experience_years as number,
      license_number: this.form.license_number.trim(),
      license_expiry_date: this.form.license_expiry_date,
      verification_status: this.form.verification_status,
      verification_notes: this.form.verification_notes.trim(),
      status: this.form.status,
      hire_date: this.form.hire_date
    };

    this.conductorService.addConductor(newConductor);
    this.isSubmitting = false;
    this.router.navigate(['/conductor-management']);
  }

  cancel(): void {
    this.router.navigate(['/conductor-management']);
  }

  goBack(): void {
    this.router.navigate(['/conductor-management']);
  }
}
