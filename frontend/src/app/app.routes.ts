import { Routes } from '@angular/router';
import { HomeComponent } from './pages/home/home.component';
import { LoginComponent } from './pages/login/login.component';
import { DashboardComponent } from './pages/dashboard/dashboard.component';
import { BusManagementComponent } from './pages/bus-management/bus-management.component';
import { BusOwnersComponent } from './pages/bus-owners/bus-owners.component';
import { EditBusComponent } from './pages/bus-management/edit-bus.component';
import { DriverManagementComponent } from './pages/driver-management/driver-management.component';
import { AddDriverComponent } from './pages/driver-management/add-driver.component';
import { EditDriverComponent } from './pages/driver-management/edit-driver.component';
import { PassengerManagementComponent } from './pages/passenger-management/passenger-management.component';
import { AddPassengerComponent } from './pages/passenger-management/add-passenger.component';
import { EditPassengerComponent } from './pages/passenger-management/edit-passenger.component';
import { LoungesManagementComponent } from './pages/lounge-management/lounges-management.component';
// import { AddLoungeComponent } from './pages/lounge-management/add-lounge.component';
// import { EditLoungeComponent } from './pages/lounge-management/edit-lounge.component';
// import { ViewLoungeComponent } from './pages/lounge-management/view-lounge.component';
import { LoungeBookingComponent } from './pages/lounge-booking/lounge-booking.component';
import { EditLoungeBookingComponent } from './pages/lounge-booking/edit-lounge-booking.component';
import { BusBookingComponent } from './pages/bus-booking/bus-booking.component';
import { ConductorManagementComponent } from './pages/conductor-management/conductor-management.component';
import { AddConductorComponent } from './pages/conductor-management/add-conductor.component';
import { EditConductorComponent } from './pages/conductor-management/edit-conductor.component';
import { NotificationDetailsComponent } from './pages/notification-details/notification-details.component';
import { SeatLayoutsComponent } from './pages/seat-layouts/seat-layouts.component';
import { SettingsComponent } from './pages/settings/settings.component';
import { ComplaintManagementComponent } from './pages/complaint-management/complaint-management.component';
import { AssignedComplaintsComponent } from './pages/assigned-complaints/assigned-complaints.component';
import { SearchResultsComponent } from './pages/search-results/search-results.component';
import { ResetPasswordComponent } from './pages/reset-password/reset-password.component';

export const routes: Routes = [
  { path: '', component: HomeComponent },
  { path: 'home', component: HomeComponent },
  { path: 'login', component: LoginComponent },
  { path: 'dashboard', component: DashboardComponent },
  { path: 'settings', component: SettingsComponent },
  { path: 'complaints', component: ComplaintManagementComponent },
  { path: 'complaints/assigned', component: AssignedComplaintsComponent },
  { path: 'bus-management', component: BusManagementComponent },
  { path: 'bus-owners', component: BusOwnersComponent },
  { path: 'bus-management/edit/:id', component: EditBusComponent },
  { path: 'driver-management', component: DriverManagementComponent },
  { path: 'driver-management/add', component: AddDriverComponent },
  { path: 'driver-management/edit/:id', component: EditDriverComponent },
  { path: 'conductor-management', component: ConductorManagementComponent },
  { path: 'conductor-management/add', component: AddConductorComponent },
  { path: 'conductor-management/edit/:id', component: EditConductorComponent },
  { path: 'passenger-management', component: PassengerManagementComponent },
  { path: 'passenger-management/add', component: AddPassengerComponent },
  { path: 'passenger-management/edit/:id', component: EditPassengerComponent },
  { path: 'lounges-management', component: LoungesManagementComponent },
  // { path: 'lounges-management/view/:id', component: ViewLoungeComponent },
  // { path: 'add-lounge', component: AddLoungeComponent },
  // { path: 'lounges-management/edit/:id', component: EditLoungeComponent },
  // { path: 'lounges-management/view/:id', component: ViewLoungeComponent },
  { path: 'lounge-booking', component: LoungeBookingComponent },
  { path: 'lounge-booking/edit/:id', component: EditLoungeBookingComponent },
  { path: 'bus-booking', component: BusBookingComponent },
  { path: 'seat-layouts', component: SeatLayoutsComponent },
  { path: 'notification-details', component: NotificationDetailsComponent },
  { path: 'search-results', component: SearchResultsComponent },
  { path: 'reset-password', component: ResetPasswordComponent },
  { path: '**', redirectTo: '' }
];
