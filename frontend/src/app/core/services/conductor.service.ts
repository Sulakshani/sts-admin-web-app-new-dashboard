import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable, tap } from 'rxjs';
import { Conductor } from '../models/conductor.model';
import { environment } from '../../../environments/environment';

@Injectable({ providedIn: 'root' })
export class ConductorService {
  private apiUrl = `${environment.apiUrl}/conductors`;
  private readonly _conductors$ = new BehaviorSubject<Conductor[]>([]);

  readonly conductors$ = this._conductors$.asObservable();

  constructor(private http: HttpClient) {
    this.loadConductors();
  }

  loadConductors(): void {
    this.http.get<Conductor[]>(this.apiUrl).subscribe({
      next: (conductors) => this._conductors$.next(conductors),
      error: (err) => console.error('Failed to load conductors', err)
    });
  }

  get conductors(): Conductor[] {
    return this._conductors$.getValue();
  }

  addConductor(conductor: Conductor): Observable<Conductor> {
    return this.http.post<Conductor>(this.apiUrl, conductor).pipe(
      tap((newConductor) => {
        this._conductors$.next([...this.conductors, newConductor]);
      })
    );
  }

  updateConductor(updated: Conductor): Observable<Conductor> {
    return this.http.put<Conductor>(`${this.apiUrl}/${updated.id}`, updated).pipe(
      tap((updatedConductor) => {
        this._conductors$.next(this.conductors.map(c => (c.id === updatedConductor.id ? updatedConductor : c)));
      })
    );
  }

  deleteConductor(conductorId: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${conductorId}`).pipe(
      tap(() => {
        this._conductors$.next(this.conductors.filter(c => c.id !== conductorId));
      })
    );
  }

  getById(conductorId: string): Conductor | undefined {
    return this.conductors.find(c => c.id === conductorId);
  }

  // Helper methods for status filtering
  getActiveConductors(): Conductor[] {
    return this.conductors.filter(c => c.status.toLowerCase() === 'active');
  }

  getOnLeaveConductors(): Conductor[] {
    return this.conductors.filter(c => c.status.toLowerCase() === 'on leave');
  }

  getResignedConductors(): Conductor[] {
    return this.conductors.filter(c => c.status === 'Resigned');
  }
}
