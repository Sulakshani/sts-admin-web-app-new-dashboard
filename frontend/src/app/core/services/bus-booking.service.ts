import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable } from 'rxjs';
import { tap, switchMap, map } from 'rxjs/operators';
import { BusBooking } from '../models/bus-booking.model';

@Injectable({ providedIn: 'root' })
export class BusBookingService {
  private readonly apiUrl = 'http://localhost:8083/api/bookings';
  private readonly _bookings$ = new BehaviorSubject<BusBooking[]>([]);

  readonly bookings$ = this._bookings$.asObservable();

  constructor(private http: HttpClient) {
    this.loadBookings().subscribe({
      error: (err) => console.error('Error loading initial bookings:', err)
    });
  }

  get bookings(): BusBooking[] { return this._bookings$.getValue(); }

  loadBookings(): Observable<BusBooking[]> {
    return this.http.get<BusBooking[]>(this.apiUrl).pipe(
      tap(bookings => {
        console.log('Loaded bookings:', bookings.length);
        this._bookings$.next(bookings);
      })
    );
  }

  add(b: BusBooking): Observable<BusBooking> {
    return this.http.post<BusBooking>(this.apiUrl, b).pipe(
      tap(() => console.log('Booking created successfully')),
      switchMap((result) => {
        // Wait for reload to complete before returning
        return this.loadBookings().pipe(
          map(() => {
            console.log('Bookings reloaded after add');
            return result;
          })
        );
      })
    );
  }

  update(b: BusBooking): Observable<BusBooking> {
    return this.http.put<BusBooking>(`${this.apiUrl}/${b.booking_id}`, b).pipe(
      tap((updatedBooking) => console.log('Booking updated successfully', updatedBooking)),
      switchMap((result) => {
        // Wait for reload to complete before returning
        return this.loadBookings().pipe(
          map(() => {
            console.log('Bookings reloaded after update');
            return result;
          })
        );
      })
    );
  }

  delete(id: string) { this._bookings$.next(this.bookings.filter(x => x.booking_id !== id)); }

  // Aggregations for charts
  countByPaymentStatus() {
    const map: Record<string, number> = { paid: 0, pending: 0, failed: 0, refunded: 0, collect_on_bus: 0 };
    this.bookings.forEach(b => {
      const status = b.payment_status;
      if (status) {
        map[status] = (map[status] || 0) + 1;
      }
    });
    return map;
  }

  countByBookingStatus() {
    const map: Record<string, number> = { confirmed: 0, pending: 0, cancelled: 0, completed: 0 };
    this.bookings.forEach(b => {
      const status = b.booking_status;
      if (status) {
        map[status] = (map[status] || 0) + 1;
      }
    });
    return map;
  }

  monthlyRevenue(year: number) {
    const arr = Array(12).fill(0);
    this.bookings
      .filter(b => new Date(b.created_at).getFullYear() === year && b.payment_status === 'paid')
      .forEach(b => {
        const m = new Date(b.created_at).getMonth();
        arr[m] += b.total_fare;
      });
    return arr;
  }
}