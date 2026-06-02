import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import { tap, switchMap } from 'rxjs/operators';
import { LoungeBooking } from '../models/lounge-booking.model';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../../environments/environment';

@Injectable({ providedIn: 'root' })
export class LoungeBookingService {
  private readonly _bookings$ = new BehaviorSubject<LoungeBooking[]>([]);
  readonly bookings$ = this._bookings$.asObservable();
  private apiUrl = `${environment.apiUrl}/lounge-bookings`;

  constructor(private http: HttpClient) {
    this.loadBookings().subscribe();
  }

  get bookings(): LoungeBooking[] { 
    return this._bookings$.getValue(); 
  }

  loadBookings(): Observable<LoungeBooking[]> {
    return this.http.get<LoungeBooking[]>(this.apiUrl).pipe(
      tap(data => {
        console.log('Loaded lounge bookings:', data.length);
        this._bookings$.next(data);
      })
    );
  }

  add(b: LoungeBooking): Observable<any> {
    return this.http.post(this.apiUrl, b).pipe(
      switchMap(() => this.loadBookings())
    );
  }

  update(b: LoungeBooking): Observable<any> {
    return this.http.put(`${this.apiUrl}/${b.lounge_booking_id}`, b).pipe(
      switchMap(() => this.loadBookings())
    );
  }

  updatePaymentStatus(id: string, status: string): Observable<any> {
    return this.http.patch(`${this.apiUrl}/${id}/payment-status`, { status }).pipe(
      switchMap(() => this.loadBookings())
    );
  }

  updateBookingStatus(id: string, status: string): Observable<any> {
    return this.http.patch(`${this.apiUrl}/${id}/booking-status`, { status }).pipe(
      switchMap(() => this.loadBookings())
    );
  }

  delete(id: string): Observable<any> {
    return this.http.delete(`${this.apiUrl}/${id}`).pipe(
      switchMap(() => this.loadBookings())
    );
  }

  getById(id: string): Observable<LoungeBooking> {
    return this.http.get<LoungeBooking>(`${this.apiUrl}/${id}`);
  }

  // Aggregations for charts
  countByPaymentStatus() {
    const map: Record<string, number> = { paid: 0, pending: 0, failed: 0 };
    this.bookings.forEach(b => map[b.payment_status] = (map[b.payment_status] || 0) + 1);
    return map;
  }

  countByBookingStatus() {
    const map: Record<string, number> = { confirmed: 0, pending: 0, cancelled: 0, completed: 0 };
    this.bookings.forEach(b => map[b.status] = (map[b.status] || 0) + 1);
    return map;
  }

  monthlyRevenue(year: number) {
    const arr = Array(12).fill(0);
    this.bookings
      .filter(b => new Date(b.scheduled_arrival).getFullYear() === year && b.payment_status === 'paid')
      .forEach(b => {
        const m = new Date(b.scheduled_arrival).getMonth();
        arr[m] += b.total_amount;
      });
    return arr;
  }
}