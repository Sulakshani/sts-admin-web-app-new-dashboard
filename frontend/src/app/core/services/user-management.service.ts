import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { AdminUser } from '../models/admin-user.model';
import { environment } from '../../../environments/environment';

export interface CreateUserRequest {
  email: string;
  full_name: string;
  password: string;
  role: 'admin' | 'supervisor' | 'super_admin';
  app_scope?: 'bus' | 'driver' | 'lounges' | 'passenger';
  supervisor_id?: string;
  permissions?: string[];
}

export interface UpdateUserRequest {
  full_name?: string;
  is_active?: boolean;
  role?: 'admin' | 'supervisor' | 'super_admin';
  app_scope?: 'bus' | 'driver' | 'lounges' | 'passenger';
  supervisor_id?: string;
  permissions?: string[];
}

export interface UsersListResponse {
  users: AdminUser[];
  total: number;
}

@Injectable({
  providedIn: 'root'
})
export class UserManagementService {
  private readonly API_URL = `${environment.apiUrl}/admin/users`;

  constructor(private http: HttpClient) {}

  createUser(userData: CreateUserRequest): Observable<AdminUser> {
    return this.http.post<AdminUser>(this.API_URL, userData);
  }

  getAllUsers(): Observable<UsersListResponse> {
    return this.http.get<UsersListResponse>(this.API_URL);
  }

  updateUser(userId: string, userData: UpdateUserRequest): Observable<AdminUser> {
    return this.http.put<AdminUser>(`${this.API_URL}/${userId}`, userData);
  }

  deleteUser(userId: string): Observable<any> {
    return this.http.delete(`${this.API_URL}/${userId}`);
  }
}
