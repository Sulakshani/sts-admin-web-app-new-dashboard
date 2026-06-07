import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable, tap } from 'rxjs';
import { Driver } from '../models/driver.model';
import { environment } from '../../../environments/environment';

@Injectable({ providedIn: 'root' })
export class DriverService {
  private apiUrl = `${environment.apiUrl}/drivers`;
  private readonly _drivers$ = new BehaviorSubject<Driver[]>([]);

  readonly drivers$ = this._drivers$.asObservable();

  constructor(private http: HttpClient) {
    this.loadDrivers();
  }

  loadDrivers(): void {
    this.http.get<Driver[]>(this.apiUrl).subscribe({
      next: (drivers) => this._drivers$.next(drivers),
      error: (err) => console.error('Failed to load drivers', err)
    });
  }

  get drivers(): Driver[] {
    return this._drivers$.getValue();
  }

  addDriver(driver: Driver): Observable<Driver> {
    return this.http.post<Driver>(this.apiUrl, driver).pipe(
      tap((newDriver) => {
        this._drivers$.next([...this.drivers, newDriver]);
      })
    );
  }

  updateDriver(updated: Driver): Observable<Driver> {
    return this.http.put<Driver>(`${this.apiUrl}/${updated.id}`, updated).pipe(
      tap((updatedDriver) => {
        this._drivers$.next(this.drivers.map(d => (d.id === updatedDriver.id ? updatedDriver : d)));
      })
    );
  }

  deleteDriver(driverId: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${driverId}`).pipe(
      tap(() => {
        this._drivers$.next(this.drivers.filter(d => d.id !== driverId));
      })
    );
  }

  getById(driverId: string): Driver | undefined {
    return this.drivers.find(d => d.id === driverId);
  }
}