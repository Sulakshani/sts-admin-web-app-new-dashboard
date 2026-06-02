import { RenderMode, ServerRoute } from '@angular/ssr';

export const serverRoutes: ServerRoute[] = [
  {
    path: 'bus-management/edit/:id',
    renderMode: RenderMode.Server
  },
  {
    path: 'driver-management/edit/:id',
    renderMode: RenderMode.Server
  },
  {
    path: 'conductor-management/edit/:id',
    renderMode: RenderMode.Server
  },
  {
    path: 'passenger-management/edit/:id',
    renderMode: RenderMode.Server
  },
  {
    path: 'lounges-management/view/:id',
    renderMode: RenderMode.Server
  },
  {
    path: 'lounges-management/edit/:id',
    renderMode: RenderMode.Server
  },
  {
    path: 'lounge-booking/edit/:id',
    renderMode: RenderMode.Server
  },
  {
    path: '**',
    renderMode: RenderMode.Prerender
  }
];
