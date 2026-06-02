import { ChangeDetectorRef, Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { AdminAuthService } from '../../core/services/admin-auth.service';
import { finalize, timeout } from 'rxjs';

@Component({
  selector: 'app-reset-password',
  standalone: true,
  imports: [FormsModule, CommonModule],
  templateUrl: './reset-password.component.html',
  styleUrls: ['./reset-password.component.scss']
})
export class ResetPasswordComponent implements OnInit {
  token: string = '';
  newPassword: string = '';
  confirmPassword: string = '';
  isLoading: boolean = false;
  successMessage: string = '';
  errorMessage: string = '';
  tokenValid: boolean = false;
  checkingToken: boolean = true;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private adminAuthService: AdminAuthService,
    private cdr: ChangeDetectorRef
  ) {}

  ngOnInit() {
    // Read token once from URL to avoid repeated verification calls.
    this.token = this.route.snapshot.queryParamMap.get('token') || '';

    if (this.token) {
      this.verifyToken();
    } else {
      this.checkingToken = false;
      this.errorMessage = 'Invalid reset link. Please request a new password reset.';
    }
  }

  verifyToken() {
    this.checkingToken = true;
    this.adminAuthService.verifyResetToken(this.token).pipe(
      timeout(10000),
      finalize(() => {
        this.checkingToken = false;
        this.cdr.detectChanges();
      })
    ).subscribe({
      next: (response) => {
        this.tokenValid = true;
        this.cdr.detectChanges();
      },
      error: (error) => {
        this.tokenValid = false;
        if (error.name === 'TimeoutError') {
          this.errorMessage = 'Verification timed out. Please check your connection and try opening the reset link again.';
        } else if (error.error?.error) {
          this.errorMessage = error.error.error;
        } else {
          this.errorMessage = 'Invalid or expired reset link. Please request a new password reset.';
        }
        this.cdr.detectChanges();
      }
    });
  }

  onResetPassword() {
    this.errorMessage = '';
    this.successMessage = '';

    // Validation
    if (!this.newPassword || !this.confirmPassword) {
      this.errorMessage = 'Please fill in all fields';
      return;
    }

    if (this.newPassword.length < 8) {
      this.errorMessage = 'Password must be at least 8 characters long';
      return;
    }

    if (this.newPassword !== this.confirmPassword) {
      this.errorMessage = 'Passwords do not match';
      return;
    }

    this.isLoading = true;

    this.adminAuthService.resetPassword(this.token, this.newPassword).pipe(
      timeout(15000),
      finalize(() => {
        this.isLoading = false;
        this.cdr.detectChanges();
      })
    ).subscribe({
      next: (response) => {
        this.successMessage = 'Password reset successfully! Redirecting to login...';
        this.cdr.detectChanges();
        
        // Redirect to login after 2 seconds
        setTimeout(() => {
          this.router.navigate(['/login']);
        }, 2000);
      },
      error: (error) => {
        console.error('❌ Password reset failed:', error);
        
        if (error.name === 'TimeoutError') {
          this.errorMessage = 'Request timed out. Please try again.';
        } else if (error.error?.error) {
          this.errorMessage = error.error.error;
        } else if (error.status === 0) {
          this.errorMessage = 'Cannot connect to server. Please check your connection.';
        } else {
          this.errorMessage = 'Failed to reset password. Please try again.';
        }
        this.cdr.detectChanges();
      }
    });
  }

  goBack() {
    this.router.navigate(['/login']);
  }
}
