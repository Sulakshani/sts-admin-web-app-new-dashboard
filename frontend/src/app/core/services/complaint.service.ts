import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { BehaviorSubject, Observable, tap } from 'rxjs';
import { Complaint, ComplaintEscalation, PaginatedComplaintsResponse } from '../models/complaint.model';
import { environment } from '../../../environments/environment';

@Injectable({ providedIn: 'root' })
export class ComplaintService {
  private apiUrl = `${environment.apiUrl}/complaints`;
  private escalationApiUrl = `${environment.apiUrl}/escalation`;
  private readonly _complaints$ = new BehaviorSubject<Complaint[]>([]);

  readonly complaints$ = this._complaints$.asObservable();

  constructor(private http: HttpClient) {}

  get complaints(): Complaint[] {
    return this._complaints$.getValue();
  }

  loadComplaints(page: number = 1, pageSize: number = 20, role: string = ''): Observable<PaginatedComplaintsResponse> {
    let params = new HttpParams()
      .set('page', String(page))
      .set('page_size', String(pageSize));

    if (role) {
      params = params.set('role', role);
    }

    return this.http.get<PaginatedComplaintsResponse>(this.apiUrl, { params }).pipe(
      tap(response => {
        console.log('Loaded complaints page:', response.page, 'count:', response.data.length, 'total:', response.total);
        this._complaints$.next(response.data);
      })
    );
  }

  loadComplaintsByRole(role: string): Observable<Complaint[]> {
    return this.http.get<Complaint[]>(`${this.apiUrl}?role=${role}`).pipe(
      tap(complaints => {
        console.log(`Loaded ${role} complaints:`, complaints.length);
      })
    );
  }

  getComplaintById(id: string): Observable<Complaint> {
    return this.http.get<Complaint>(`${this.apiUrl}/${id}`);
  }

  updateComplaintStatus(id: string, status: string, resolutionNotes?: string): Observable<any> {
    return this.http.put(`${this.apiUrl}/${id}/status`, {
      status,
      resolution_notes: resolutionNotes
    }).pipe(
      tap(() => this.loadComplaints(1, 20).subscribe())
    );
  }

  // Escalation methods
  getComplaintEscalation(id: string): Observable<ComplaintEscalation> {
    return this.http.get<ComplaintEscalation>(`${this.apiUrl}/${id}/escalation`);
  }

  manualEscalateComplaint(id: string, escalatedBy: string = 'admin'): Observable<any> {
    return this.http.post(`${this.apiUrl}/${id}/escalate`, {
      escalated_by: escalatedBy
    }).pipe(
      tap(() => this.loadComplaints(1, 20).subscribe())
    );
  }

  assignComplaint(id: string, appRole: string, appScope?: string): Observable<any> {
    return this.http.post(`${this.escalationApiUrl}/complaint/${id}/assign`, {
      app_role: appRole,
      app_scope: appScope || ''
    }).pipe(
      tap(() => this.loadComplaints(1, 20).subscribe())
    );
  }

  getEscalationStats(): Observable<any> {
    return this.http.get(`${this.escalationApiUrl}/stats`);
  }

  initializeEscalation(complaintId: string, category: string): Observable<any> {
    return this.http.post(`${this.escalationApiUrl}/complaint/${complaintId}/initialize`, {
      category
    });
  }
}
