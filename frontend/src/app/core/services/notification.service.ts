import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable, forkJoin, of } from 'rxjs';
import { catchError, tap, map } from 'rxjs/operators';

export interface BusNotification {
  id: string;
  bus_number: string;
  company_name: string;
  identify_or_incorporation_no: string;
  business_email: string;
  business_phone: string;
  permit_number: string;
  license_plate: string;
  total_seats: number;
  bus_type: string;
  custom_route_name: string;
  fare_per_seat: number;
  status: string;
  verification_status: string;
  verification_documents?: string[];
  created_at?: string;
}

export interface DriverNotification {
  id: string;
  name: string;
  contact_number: string;
  license_number: string;
  license_expiry_date: string;
  experience_years: number;
  verification_status: string;
  verification_notes: string;
  status: string;
  hire_date: string;
  created_at?: string;
}

export interface ConductorNotification {
  id: string;
  name: string;
  contact_number: string;
  license_number: string;
  license_expiry_date: string;
  experience_years: number;
  verification_status: string;
  verification_notes: string;
  status: string;
  hire_date: string;
  created_at?: string;
}

export interface LoungeNotification {
  lounge_id: string;
  lounge_owner: string;
  owner_nic: string;
  owner_email: string;
  owner_contact: string;
  lounge_name: string;
  lounge_contact: string;
  address: string;
  capacity: number;
  price_per_hour: number;
  facilities: string[];
  marketplace: string;
  verification: string;
  verification_note: string;
  operational: boolean;
  created_at?: string;
}

export interface LoungeOwnerNotification {
  id: string;
  user_id: string;
  manager_full_name: string;
  email: string;
  contact_number: string;
  nic: string;
  business_name: string;
  business_license: string;
  verification_status: string;
  verification_notes: string;
  created_at?: string;
  updated_at?: string;
}

export interface BusOwnerNotification {
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

export type AllNotifications =
  | BusNotification
  | DriverNotification
  | ConductorNotification
  | LoungeNotification
  | BusOwnerNotification
  | LoungeOwnerNotification;

@Injectable({ providedIn: 'root' })
export class NotificationService {
  private readonly apiUrl = 'http://localhost:8083/api';
  private readonly _pendingBuses$ = new BehaviorSubject<BusNotification[]>([]);
  private readonly _pendingDrivers$ = new BehaviorSubject<DriverNotification[]>([]);
  private readonly _pendingConductors$ = new BehaviorSubject<ConductorNotification[]>([]);
  private readonly _pendingLounges$ = new BehaviorSubject<LoungeNotification[]>([]);
  private readonly _pendingBusOwners$ = new BehaviorSubject<BusOwnerNotification[]>([]);
  private readonly _pendingLoungeOwners$ = new BehaviorSubject<LoungeOwnerNotification[]>([]);

  readonly pendingBuses$ = this._pendingBuses$.asObservable();
  readonly pendingDrivers$ = this._pendingDrivers$.asObservable();
  readonly pendingConductors$ = this._pendingConductors$.asObservable();
  readonly pendingLounges$ = this._pendingLounges$.asObservable();
  readonly pendingBusOwners$ = this._pendingBusOwners$.asObservable();
  readonly pendingLoungeOwners$ = this._pendingLoungeOwners$.asObservable();

  constructor(private http: HttpClient) {
    this.loadAllPendingNotifications();
  }

  get totalPendingCount(): number {
    return this._pendingBuses$.getValue().length +
           this._pendingDrivers$.getValue().length +
           this._pendingConductors$.getValue().length +
           this._pendingLounges$.getValue().length +
           this._pendingBusOwners$.getValue().length +
           this._pendingLoungeOwners$.getValue().length;
  }

  loadAllPendingNotifications(): void {
    forkJoin({
      buses: this.http.get<BusNotification[]>(`${this.apiUrl}/buses/pending`).pipe(
        catchError((err) => {
          console.error('Error loading pending buses:', err);
          return of([] as BusNotification[]);
        })
      ),
      drivers: this.http.get<DriverNotification[]>(`${this.apiUrl}/drivers/pending`).pipe(
        catchError((err) => {
          console.error('Error loading pending drivers:', err);
          return of([] as DriverNotification[]);
        })
      ),
      conductors: this.http.get<ConductorNotification[]>(`${this.apiUrl}/conductors/pending`).pipe(
        catchError((err) => {
          console.error('Error loading pending conductors:', err);
          return of([] as ConductorNotification[]);
        })
      ),
      lounges: this.http.get<LoungeNotification[]>(`${this.apiUrl}/lounges/pending`).pipe(
        catchError((err) => {
          console.error('Error loading pending lounges:', err);
          return of([] as LoungeNotification[]);
        })
      ),
      busOwners: this.http.get<BusOwnerNotification[]>(`${this.apiUrl}/bus-owners/pending`).pipe(
        catchError((err) => {
          console.error('Error loading pending bus owners:', err);
          return of([] as BusOwnerNotification[]);
        })
      ),
      loungeOwners: this.http.get<LoungeOwnerNotification[]>(`${this.apiUrl}/lounge-owners/pending`).pipe(
        catchError((err) => {
          console.error('Error loading pending lounge owners:', err);
          return of([] as LoungeOwnerNotification[]);
        })
      )
    }).subscribe({
      next: (data) => {
        this._pendingBuses$.next(data.buses);
        this._pendingDrivers$.next(data.drivers);
        this._pendingConductors$.next(data.conductors);
        this._pendingLounges$.next(data.lounges);
        this._pendingBusOwners$.next(data.busOwners);
        this._pendingLoungeOwners$.next(data.loungeOwners);
        console.log('Loaded pending notifications:', {
          buses: data.buses.length,
          drivers: data.drivers.length,
          conductors: data.conductors.length,
          lounges: data.lounges.length,
          busOwners: data.busOwners.length,
          loungeOwners: data.loungeOwners.length
        });
      },
      error: (err) => console.error('Error loading pending notifications:', err)
    });
  }

  // Bus methods
  approveBus(busId: string, data?: any): Observable<any> {
    const requestBody = { status: 'Verified', ...data };
    return this.http.put(`${this.apiUrl}/buses/${busId}/verify`, requestBody).pipe(
      tap(() => this.loadAllPendingNotifications())
    );
  }

  rejectBus(busId: string, data?: any): Observable<any> {
    const requestBody = { status: 'Rejected', ...data };
    return this.http.put(`${this.apiUrl}/buses/${busId}/verify`, requestBody).pipe(
      tap(() => this.loadAllPendingNotifications())
    );
  }

  getBusById(busId: string): Observable<BusNotification> {
    return this.http.get<BusNotification>(`${this.apiUrl}/buses/${busId}`);
  }

  // Driver methods
  approveDriver(driverId: string, data?: any): Observable<any> {
    const requestBody = { status: 'Verified', ...data };
    return this.http.put(`${this.apiUrl}/drivers/${driverId}/verify`, requestBody).pipe(
      tap(() => this.loadAllPendingNotifications())
    );
  }

  rejectDriver(driverId: string, data?: any): Observable<any> {
    const requestBody = { status: 'Rejected', ...data };
    return this.http.put(`${this.apiUrl}/drivers/${driverId}/verify`, requestBody).pipe(
      tap(() => this.loadAllPendingNotifications())
    );
  }

  getDriverById(driverId: string): Observable<DriverNotification> {
    return this.http.get<DriverNotification>(`${this.apiUrl}/drivers/${driverId}`);
  }

  // Conductor methods
  approveConductor(conductorId: string, data?: any): Observable<any> {
    const requestBody = { status: 'Verified', ...data };
    return this.http.put(`${this.apiUrl}/conductors/${conductorId}/verify`, requestBody).pipe(
      tap(() => this.loadAllPendingNotifications())
    );
  }

  rejectConductor(conductorId: string, data?: any): Observable<any> {
    const requestBody = { status: 'Rejected', ...data };
    return this.http.put(`${this.apiUrl}/conductors/${conductorId}/verify`, requestBody).pipe(
      tap(() => this.loadAllPendingNotifications())
    );
  }

  getConductorById(conductorId: string): Observable<ConductorNotification> {
    return this.http.get<ConductorNotification>(`${this.apiUrl}/conductors/${conductorId}`);
  }

  // Lounge methods
  approveLounge(loungeId: string, data?: any): Observable<any> {
    const requestBody = { status: 'Verified', ...data };
    return this.http.put(`${this.apiUrl}/lounges/${loungeId}/verify`, requestBody).pipe(
      tap(() => this.loadAllPendingNotifications())
    );
  }

  rejectLounge(loungeId: string, data?: any): Observable<any> {
    const requestBody = { status: 'Rejected', ...data };
    return this.http.put(`${this.apiUrl}/lounges/${loungeId}/verify`, requestBody).pipe(
      tap(() => this.loadAllPendingNotifications())
    );
  }

  getLoungeById(loungeId: string): Observable<LoungeNotification> {
    return this.http.get<LoungeNotification>(`${this.apiUrl}/lounges/${loungeId}`);
  }

  // Bus Owner methods
  approveBusOwner(busOwnerId: string, data?: any): Observable<any> {
    const requestBody = { verification_status: 'verified', ...data };
    return this.http.put(`${this.apiUrl}/bus-owners/${busOwnerId}/verify`, requestBody).pipe(
      tap(() => this.loadAllPendingNotifications())
    );
  }

  rejectBusOwner(busOwnerId: string, data?: any): Observable<any> {
    const requestBody = { verification_status: 'rejected', ...data };
    return this.http.put(`${this.apiUrl}/bus-owners/${busOwnerId}/verify`, requestBody).pipe(
      tap(() => this.loadAllPendingNotifications())
    );
  }

  getBusOwnerById(busOwnerId: string): Observable<BusOwnerNotification> {
    return this.http.get<BusOwnerNotification>(`${this.apiUrl}/bus-owners/${busOwnerId}`);
  }

  // Lounge Owner methods
  approveLoungeOwner(loungeOwnerId: string, data?: any): Observable<any> {
    const requestBody = { verification_status: 'approved', ...data };
    return this.http.put(`${this.apiUrl}/lounge-owners/${loungeOwnerId}/verify`, requestBody).pipe(
      tap(() => this.loadAllPendingNotifications())
    );
  }

  rejectLoungeOwner(loungeOwnerId: string, data?: any): Observable<any> {
    const requestBody = { verification_status: 'rejected', ...data };
    return this.http.put(`${this.apiUrl}/lounge-owners/${loungeOwnerId}/verify`, requestBody).pipe(
      tap(() => this.loadAllPendingNotifications())
    );
  }

  getLoungeOwnerById(loungeOwnerId: string): Observable<LoungeOwnerNotification> {
    return this.http.get<LoungeOwnerNotification>(`${this.apiUrl}/lounge-owners/${loungeOwnerId}`);
  }
}
