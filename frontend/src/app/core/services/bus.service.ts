import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable, tap, switchMap } from 'rxjs';
import { Bus } from '../models/bus.model';
import { environment } from '../../../environments/environment';

@Injectable({ providedIn: 'root' })
export class BusService {
  private apiUrl = `${environment.apiUrl}/buses`;
  private readonly _buses$ = new BehaviorSubject<Bus[]>([]);
  readonly buses$ = this._buses$.asObservable();

  constructor(private http: HttpClient) {
    this.loadBuses().subscribe();
  }

  get buses(): Bus[] {
    return this._buses$.getValue();
  }

  loadBuses(): Observable<Bus[]> {
    return this.http.get<Bus[]>(this.apiUrl).pipe(
      tap(buses => {
        console.log('Loaded buses:', buses.length);
        this._buses$.next(buses);
      })
    );
  }

  addBus(bus: Omit<Bus, 'id'>): Observable<Bus[]> {
    return this.http.post<Bus>(this.apiUrl, bus).pipe(
      switchMap(() => this.loadBuses()),
      tap(() => console.log('Bus added and list reloaded'))
    );
  }

  updateBus(updated: Bus): Observable<Bus[]> {
    return this.http.put(`${this.apiUrl}/${updated.id}`, updated).pipe(
      switchMap(() => this.loadBuses())
    );
  }

  deleteBus(busId: string): Observable<Bus[]> {
    return this.http.delete(`${this.apiUrl}/${busId}`).pipe(
      switchMap(() => this.loadBuses())
    );
  }

  getById(busId: string): Bus | undefined {
    return this.buses.find(b => b.id === busId);
  }
}
