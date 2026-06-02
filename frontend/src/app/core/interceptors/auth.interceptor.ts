import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, switchMap, throwError } from 'rxjs';
import { AdminAuthService } from '../services/admin-auth.service';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const authService = inject(AdminAuthService);
  const router = inject(Router);
  const token = authService.getAccessToken();

  // Clone the request and add the authorization header if token exists
  const clonedReq = token ? req.clone({
    setHeaders: {
      Authorization: `Bearer ${token}`
    }
  }) : req;

  return next(clonedReq).pipe(
    catchError((error: HttpErrorResponse) => {
      // If 401 error and not already on auth routes
      if (error.status === 401 && !req.url.includes('/auth/')) {
        // Try to refresh the token
        return authService.refreshToken().pipe(
          switchMap(() => {
            // Retry the original request with new token
            const newToken = authService.getAccessToken();
            const retryReq = req.clone({
              setHeaders: {
                Authorization: `Bearer ${newToken}`
              }
            });
            return next(retryReq);
          }),
          catchError((refreshError) => {
            // Refresh failed, redirect to login
            authService.logout().subscribe({
              complete: () => router.navigate(['/login'])
            });
            return throwError(() => refreshError);
          })
        );
      }

      return throwError(() => error);
    })
  );
};
