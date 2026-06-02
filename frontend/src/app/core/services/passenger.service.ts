import { Injectable } from '@angular/core';
import { BehaviorSubject } from 'rxjs';
import { Passenger } from '../models/passenger.model';

@Injectable({ providedIn: 'root' })
export class PassengerService {
  private readonly _passengers$ = new BehaviorSubject<Passenger[]>([
    { passenger_id: 'PASS001', name: 'John Doe', phone: '+94712345678', email: 'john@example.com', nic: '123456789V', created_at: '2025-01-12T12:00:00Z' },
    { passenger_id: 'PASS002', name: 'Jane Smith', phone: '+94719876543', email: 'jane@example.com', nic: '987654321V', created_at: '2025-03-22T10:00:00Z' },
    { passenger_id: 'PASS003', name: 'Bob Lee', phone: '+94711234567', email: 'bob@example.com', nic: '753159846V', created_at: '2025-05-05T09:20:00Z' }
  ]);

  readonly passengers$ = this._passengers$.asObservable();

  // For tracking bookings data
  private busBookingsData: any[] = [];
  private loungeBookingsData: any[] = [];

  get passengers(): Passenger[] { return this._passengers$.getValue(); }

  addPassenger(p: Partial<Passenger>): void {
    const existingIds = this.passengers.map(pass => parseInt(pass.passenger_id.replace('PASS', '')));
    const nextId = Math.max(...existingIds) + 1;
    const newPassenger: Passenger = {
      ...p,
      passenger_id: `PASS${nextId.toString().padStart(3, '0')}`,
      created_at: new Date().toISOString()
    } as Passenger;
    this._passengers$.next([...this.passengers, newPassenger]);
  }
  updatePassenger(updated: Passenger): void { this._passengers$.next(this.passengers.map(x => x.passenger_id === updated.passenger_id ? updated : x)); }
  deletePassenger(id: string): void { this._passengers$.next(this.passengers.filter(x => x.passenger_id !== id)); }
  getById(id: string): Passenger | undefined { return this.passengers.find(x => x.passenger_id === id); }
  getPassengerById(id: string): Passenger | undefined { return this.getById(id); }

  // Update bookings data from external sources
  setBusBookingsData(bookings: any[]): void {
    this.busBookingsData = bookings;
  }

  setLoungeBookingsData(bookings: any[]): void {
    this.loungeBookingsData = bookings;
  }

  getMonthlyCounts(year: number): number[] {
    const counts = new Array(12).fill(0);
    
    // Count bus bookings by month
    this.busBookingsData.forEach(b => {
      const d = new Date(b.created_at || b.booking_date);
      if (d.getFullYear() === year) counts[d.getMonth()]++;
    });
    
    // Count lounge bookings by month
    this.loungeBookingsData.forEach(b => {
      const d = new Date(b.created_at);
      if (d.getFullYear() === year) counts[d.getMonth()]++;
    });
    
    return counts;
  }
}
