import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterModule } from '@angular/router';

interface MenuItem {
  label: string;
  icon: string;
  route: string;
  badge?: number;
  children?: { label: string; route: string; }[];
}

@Component({
  selector: 'app-navbar',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './navbar.component.html',
  styleUrls: ['./navbar.component.scss']
})
export class NavbarComponent implements OnInit {
  menuItems: MenuItem[] = [
    { label: 'Dashboard', icon: 'fas fa-th-large', route: '/dashboard' },
    { 
      label: 'Bus', 
      icon: 'fas fa-bus', 
      route: '/bus-management',
      children: [
        { label: 'Bus Owners', route: '/bus-owners' },
        { label: 'Route Permits', route: '/bus-management' }
      ]
    },
    { label: 'Driver', icon: 'fas fa-user-tie', route: '/driver-management' },
    { label: 'Conductor', icon: 'fas fa-user-secret', route: '/conductor-management' },
    { label: 'Passenger', icon: 'fas fa-users', route: '/passenger-management' },
    { label: 'Lounge', icon: 'fas fa-couch', route: '/lounges-management' },
    { label: 'Bus Bookings', icon: 'fas fa-ticket-alt', route: '/bus-booking' },
    { label: 'Lounge Bookings', icon: 'fas fa-clipboard-list', route: '/lounge-booking' },
    { 
      label: 'Support', 
      icon: 'fas fa-headset', 
      route: '/support', 
      badge: 3,
      children: [
        { label: 'Complaints', route: '/complaints' },
        { label: 'Assigned Complaints', route: '/complaints/assigned' }
       
      ]
    },
    { label: 'Setting', icon: 'fas fa-cog', route: '/settings' }
  ];

  activeDropdown: string | null = null;

  constructor(private router: Router) {}

  ngOnInit() {}

  isActive(route: string): boolean {
    return this.router.url === route;
  }

  toggleDropdown(label: string) {
    if (this.activeDropdown === label) {
      this.activeDropdown = null;
    } else {
      this.activeDropdown = label;
    }
  }
}
