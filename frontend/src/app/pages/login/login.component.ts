import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { AdminAuthService } from '../../core/services/admin-auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [FormsModule, CommonModule],
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent {
  username: string = '';
  password: string = '';
  isLoading: boolean = false;
  errorMessage: string = '';

  // Forgot password properties
  showForgotPasswordModal: boolean = false;
  resetEmail: string = '';
  isResetLoading: boolean = false;
  resetSuccessMessage: string = '';
  resetErrorMessage: string = '';

  constructor(
    private router: Router,
    private adminAuthService: AdminAuthService
  ) {}

  onLogin() {
    if (!this.username || !this.password) {
      this.errorMessage = 'Please enter both email and password';
      return;
    }

    this.isLoading = true;
    this.errorMessage = '';

    console.log('Attempting login with:', this.username);

    // Set a timeout to prevent infinite loading
    const timeoutId = setTimeout(() => {
      if (this.isLoading) {
        this.isLoading = false;
        this.errorMessage = 'Request timeout. Server is taking too long to respond. Please try again.';
        console.error('Request timeout after 30 seconds');
      }
    }, 30000); // 30 second timeout

    // Real login using admin auth service
    this.adminAuthService.login(this.username, this.password).subscribe({
      next: (response) => {
        clearTimeout(timeoutId);
        this.isLoading = false;
        console.log('✅ Login successful', response);
        // Navigate to admin dashboard
        this.router.navigate(['/dashboard']);
      },
      error: (error) => {
        clearTimeout(timeoutId);
        this.isLoading = false;
        console.error('❌ Login failed with error:', error);
        console.error('Error status:', error.status);
        console.error('Error details:', error.error);

        // Handle different error responses with detailed messages
        if (error.status === 400) {
          this.errorMessage = 'Bad request. Please check your email and password format.';
          console.error('400 Bad Request - Request format issue');
        } else if (error.status === 401) {
          this.errorMessage = 'Invalid email or password. Please check your credentials.';
          console.error('401 Unauthorized - Invalid credentials');
        } else if (error.status === 403) {
          this.errorMessage = 'Access forbidden. Your account may be inactive.';
          console.error('403 Forbidden - Account inactive or no permission');
        } else if (error.status === 404) {
          this.errorMessage = 'Login endpoint not found. Backend may not be configured correctly.';
          console.error('404 Not Found - Endpoint missing');
        } else if (error.status === 500) {
          this.errorMessage = 'Server error. Please try again later or contact support.';
          console.error('500 Internal Server Error');
        } else if (error.status === 0) {
          this.errorMessage = 'Cannot connect to server. Check your internet connection or backend may be down.';
          console.error('Network error - CORS, network failure, or server down');
        } else if (error.error?.error) {
          this.errorMessage = error.error.error;
          console.error('Custom error message:', error.error.error);
        } else if (error.message) {
          this.errorMessage = `Error: ${error.message}`;
          console.error('Error message:', error.message);
        } else {
          this.errorMessage = `Login failed (Status: ${error.status}). Please try again or contact support.`;
          console.error('Unknown error:', error);
        }
      },
      complete: () => {
        clearTimeout(timeoutId);
        // Ensure loading is always stopped
        if (this.isLoading) {
          this.isLoading = false;
          console.log('Login request completed');
        }
      }
    });
  }

  goBack() {
    this.router.navigate(['/']);
  }

  // Forgot password methods
  openForgotPasswordModal(event: Event) {
    event.preventDefault();
    this.showForgotPasswordModal = true;
    this.resetEmail = '';
    this.resetSuccessMessage = '';
    this.resetErrorMessage = '';
  }

  closeForgotPasswordModal() {
    this.showForgotPasswordModal = false;
    this.resetEmail = '';
    this.resetSuccessMessage = '';
    this.resetErrorMessage = '';
  }

  onForgotPassword() {
    if (!this.resetEmail) {
      this.resetErrorMessage = 'Please enter your email address';
      return;
    }

    // Basic email validation
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(this.resetEmail)) {
      this.resetErrorMessage = 'Please enter a valid email address';
      return;
    }

    this.isResetLoading = true;
    this.resetErrorMessage = '';
    this.resetSuccessMessage = '';

    // Call backend API for password reset
    this.adminAuthService.requestPasswordReset(this.resetEmail).subscribe({
      next: (response: any) => {
        this.isResetLoading = false;
        this.resetSuccessMessage = response.message || 'Password reset code has been sent to your email. Please check your email for the reset code.';
        
        // In development, show the reset token
        if (response.reset_token) {
          console.log('🔑 Reset Token (DEV ONLY):', response.reset_token);
          this.resetSuccessMessage += ` (Dev Token: ${response.reset_token})`;
        }
        
        // Auto-close modal after 5 seconds
        setTimeout(() => {
          this.closeForgotPasswordModal();
        }, 5000);
      },
      error: (error) => {
        this.isResetLoading = false;
        console.error('❌ Password reset request failed:', error);
        
        if (error.error?.error) {
          this.resetErrorMessage = error.error.error;
        } else if (error.status === 0) {
          this.resetErrorMessage = 'Cannot connect to server. Please check your connection.';
        } else {
          this.resetErrorMessage = 'Failed to send reset email. Please try again.';
        }
      }
    });
  }
}
