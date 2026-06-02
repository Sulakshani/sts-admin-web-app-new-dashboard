import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable } from 'rxjs';
import { tap } from 'rxjs/operators';
import { environment } from '../../../environments/environment';

export interface BusOwner {
  id: string;
  user_id: string;
  company_name: string;
  business_email: string;
  business_phone: string;
  identity_or_incorporation_no: string;
  verification_status: string;
  verification_documents: any;
  created_at: string;
  updated_at: string;
}

@Injectable({
  providedIn: 'root'
})
export class BusOwnerService {
  private apiUrl = `${environment.apiUrl}/bus-owners`;
  private busOwnersSubject = new BehaviorSubject<BusOwner[]>([]);
  public busOwners$ = this.busOwnersSubject.asObservable();

  constructor(private http: HttpClient) {
    this.loadBusOwners();
  }

  loadBusOwners(): void {
    this.http.get<BusOwner[]>(this.apiUrl).subscribe({
      next: (data) => {
        this.busOwnersSubject.next(data);
      },
      error: (error) => {
        console.error('Error loading bus owners:', error);
      }
    });
  }

  getBusOwners(): Observable<BusOwner[]> {
    return this.http.get<BusOwner[]>(this.apiUrl);
  }

  getBusOwnerById(id: string): Observable<BusOwner> {
    return this.http.get<BusOwner>(`${this.apiUrl}/${id}`);
  }

  createBusOwner(owner: Partial<BusOwner>): Observable<BusOwner> {
    return this.http.post<BusOwner>(this.apiUrl, owner).pipe(
      tap(() => this.loadBusOwners())
    );
  }

  updateBusOwner(id: string, owner: Partial<BusOwner>): Observable<BusOwner> {
    return this.http.put<BusOwner>(`${this.apiUrl}/${id}`, owner).pipe(
      tap(() => this.loadBusOwners())
    );
  }

  deleteBusOwner(id: string): Observable<any> {
    return this.http.delete(`${this.apiUrl}/${id}`).pipe(
      tap(() => this.loadBusOwners())
    );
  }

  verifyBusOwner(id: string, data: any): Observable<any> {
    return this.http.put(`${this.apiUrl}/${id}/verify`, data).pipe(
      tap(() => this.loadBusOwners())
    );
  }
}
