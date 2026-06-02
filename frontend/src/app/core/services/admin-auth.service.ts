import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { BehaviorSubject, Observable, tap } from 'rxjs';
import { AdminUser, AdminLoginRequest, AdminLoginResponse, AdminChangePasswordRequest } from '../models/admin-user.model';
import { environment } from '../../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class AdminAuthService {
  private readonly API_URL = `${environment.apiUrl}/admin/auth`;
  private currentAdminSubject = new BehaviorSubject<AdminUser | null>(null);
  public currentAdmin$ = this.currentAdminSubject.asObservable();

  constructor(private http: HttpClient) {
    this.loadAdminFromStorage();
  }

  private loadAdminFromStorage(): void {
    const adminData = localStorage.getItem('admin_user');
    if (adminData) {
      try {
        const admin = JSON.parse(adminData);
        this.currentAdminSubject.next(admin);
      } catch (error) {
        console.error('Error parsing admin data from storage:', error);
        localStorage.removeItem('admin_user');
      }
    }
  }

  login(email: string, password: string): Observable<AdminLoginResponse> {
    const loginRequest: AdminLoginRequest = { email, password };

    return this.http.post<AdminLoginResponse>(`${this.API_URL}/login`, loginRequest)
      .pipe(
        tap(response => {
          // Store tokens
          localStorage.setItem('access_token', response.access_token);
          localStorage.setItem('refresh_token', response.refresh_token);
          localStorage.setItem('admin_user', JSON.stringify(response.admin_user));

          // Update current admin
          this.currentAdminSubject.next(response.admin_user);
        })
      );
  }

  logout(): Observable<any> {
    const refreshToken = localStorage.getItem('refresh_token');

    return this.http.post(`${this.API_URL}/logout`, { refresh_token: refreshToken })
      .pipe(
        tap(() => {
          this.clearSession();
        })
      );
  }

  private clearSession(): void {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('admin_user');
    this.currentAdminSubject.next(null);
  }

  refreshToken(): Observable<AdminLoginResponse> {
    const refreshToken = localStorage.getItem('refresh_token');

    return this.http.post<AdminLoginResponse>(`${this.API_URL}/refresh`, { refresh_token: refreshToken })
      .pipe(
        tap(response => {
          localStorage.setItem('access_token', response.access_token);
          localStorage.setItem('admin_user', JSON.stringify(response.admin_user));
          this.currentAdminSubject.next(response.admin_user);
        })
      );
  }

  getProfile(): Observable<AdminUser> {
    return this.http.get<AdminUser>(`${this.API_URL}/profile`);
  }

  changePassword(oldPassword: string, newPassword: string): Observable<any> {
    const request: AdminChangePasswordRequest = {
      old_password: oldPassword,
      new_password: newPassword
    };

    return this.http.post(`${this.API_URL}/change-password`, request);
  }

  isAuthenticated(): boolean {
    const token = localStorage.getItem('access_token');
    return !!token;
  }

  getAccessToken(): string | null {
    return localStorage.getItem('access_token');
  }

  getCurrentAdmin(): AdminUser | null {
    return this.currentAdminSubject.value;
  }

  private normalizeRole(role?: string): string {
    return (role || '').toLowerCase().trim().replace(/[-\s]+/g, '_');
  }

  isSuperAdmin(): boolean {
    const admin = this.getCurrentAdmin();
    return this.normalizeRole(admin?.role) === 'super_admin';
  }

  isAdmin(): boolean {
    const admin = this.getCurrentAdmin();
    const role = this.normalizeRole(admin?.role);
    return role === 'admin' || role === 'super_admin';
  }

  requestPasswordReset(email: string): Observable<any> {
    return this.http.post(`${this.API_URL}/forgot-password`, { email });
  }

  verifyResetToken(token: string): Observable<any> {
    return this.http.post(`${this.API_URL}/verify-reset-token`, { token });
  }

  resetPassword(token: string, newPassword: string): Observable<any> {
    return this.http.post(`${this.API_URL}/reset-password`, {
      token,
      new_password: newPassword
    });
  }
}
