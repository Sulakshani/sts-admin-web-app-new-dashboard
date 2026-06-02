import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { ConductorService } from '../../core/services/conductor.service';
import { Conductor } from '../../core/models/conductor.model';

interface ValidationErrors {
  name?: string | null;
  contact_number?: string | null;
  experience_years?: string | null;
  license_number?: string | null;
  license_expiry_date?: string | null;
  hire_date?: string | null;
}

@Component({
  selector: 'app-edit-conductor',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './edit-conductor.component.html',
  styleUrls: ['./edit-conductor.component.scss']
})
export class EditConductorComponent implements OnInit {
  conductor: Conductor | undefined;
  isSubmitting = false;
  isFormSubmitted = false;
  validationErrors: ValidationErrors = {};

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private conductorService: ConductorService
  ) {}

  ngOnInit(): void {
    const conductorId = this.route.snapshot.paramMap.get('id');
    if (conductorId) {
      this.conductor = this.conductorService.getById(conductorId);
      if (!this.conductor) {
        this.router.navigate(['/conductor-management']);
      }
    }
  }

  // Validation methods (same as add-conductor)
  private validateName(name: string): string | null {
    if (!name || name.trim() === '') return 'Name is required';
    if (name.trim().length < 2) return 'Name must be at least 2 characters';
    const pattern = /^[A-Za-z\s'-]{2,100}$/;
    if (!pattern.test(name.trim())) return 'Name can only contain letters, spaces, hyphens and apostrophes';
    return null;
  }

  private validateContactNumber(phone: string): string | null {
    if (!phone || phone.trim() === '') return 'Contact Number is required';
    const pattern = /^\d{10}$/;
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
  onNameChange(): void { if (this.conductor) this.validationErrors.name = this.validateName(this.conductor.name); }
  onContactNumberChange(): void { if (this.conductor) this.validationErrors.contact_number = this.validateContactNumber(this.conductor.contact_number); }
  onExperienceChange(): void { if (this.conductor) this.validationErrors.experience_years = this.validateExperience(this.conductor.experience_years); }
  onLicenseNumberChange(): void { if (this.conductor) this.validationErrors.license_number = this.validateLicenseNumber(this.conductor.license_number); }
  onLicenseExpiryChange(): void { if (this.conductor) this.validationErrors.license_expiry_date = this.validateLicenseExpiryDate(this.conductor.license_expiry_date); }
  onHireDateChange(): void { if (this.conductor) this.validationErrors.hire_date = this.validateHireDate(this.conductor.hire_date); }

  isFormValid(): boolean {
    if (!this.conductor) return false;
    return !this.validateName(this.conductor.name) &&
           !this.validateContactNumber(this.conductor.contact_number) &&
           !this.validateExperience(this.conductor.experience_years) &&
           !this.validateLicenseNumber(this.conductor.license_number) &&
           !this.validateLicenseExpiryDate(this.conductor.license_expiry_date) &&
           !this.validateHireDate(this.conductor.hire_date);
  }

  save(): void {
    if (!this.conductor) return;

    this.isFormSubmitted = true;
    this.validationErrors = {
      name: this.validateName(this.conductor.name),
      contact_number: this.validateContactNumber(this.conductor.contact_number),
      experience_years: this.validateExperience(this.conductor.experience_years),
      license_number: this.validateLicenseNumber(this.conductor.license_number),
      license_expiry_date: this.validateLicenseExpiryDate(this.conductor.license_expiry_date),
      hire_date: this.validateHireDate(this.conductor.hire_date)
    };

    if (!this.isFormValid()) {
      alert('Please fill all required fields correctly');
      return;
    }

    this.isSubmitting = true;
    this.conductorService.updateConductor(this.conductor).subscribe({
      next: () => {
        this.isSubmitting = false;
        alert('Successfully updated');
        this.router.navigate(['/conductor-management']);
      },
      error: (err) => {
        this.isSubmitting = false;
        console.error('Error updating conductor', err);
        alert('Failed to update conductor');
      }
    });
  }

  cancel(): void {
    this.router.navigate(['/conductor-management']);
  }

  goBack(): void {
    this.router.navigate(['/conductor-management']);
  }
}
