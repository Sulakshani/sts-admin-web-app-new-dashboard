import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { NotificationPanelComponent } from '../../shared/components/notification-panel/notification-panel.component';
import { NavbarComponent } from '../../shared/components/navbar/navbar.component';
import { BusService } from '../../core/services/bus.service';
import { LoungeService } from '../../core/services/lounge.service';
import { DriverService } from '../../core/services/driver.service';
import { ConductorService } from '../../core/services/conductor.service';
import { BusBookingService } from '../../core/services/bus-booking.service';
import { LoungeBookingService } from '../../core/services/lounge-booking.service';
import { PassengerService } from '../../core/services/passenger.service';
import { NotificationService } from '../../core/services/notification.service';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, FormsModule, NotificationPanelComponent, NavbarComponent, RouterModule],
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.scss']
})
export class DashboardComponent implements OnInit, OnDestroy {
  showNotificationPanel = false;
  private pendingNotificationsRefreshId: ReturnType<typeof setInterval> | null = null;

  numBuses = 0;
  numLounges = 0;
  numDrivers = 0;
  numConductors = 0;
  totalLoungeRevenue = 0;

  // Search functionality
  selectedSearchType: string = 'Bus';
  selectedAttributes: string[] = [];
  searchValues: { [key: string]: string } = {};
  searchResults: any[] = [];
  showSearchResults = false;

  // Attribute options for each search type
  attributeOptions: { [key: string]: string[] } = {
    Bus: ['Company', 'Route', 'Permit Num', 'Register Num', 'Owner Verification', 'Permit Verify', 'Contact', 'No of Seat', 'Approved fare', 'Type', 'Status'],
    Lounge: ['Lounge Name', 'Owner', 'Contact', 'Address', 'Price per hour', 'Capacity', 'Operation', 'Verification'],
    Driver: ['Name', 'Contact', 'License Num', 'License Expire date', 'Experience', 'Hire date', 'Verification', 'Status'],
    Conductor: ['Name', 'Contact', 'License Num', 'Experience', 'Hire date', 'Verification', 'Status'],
    'Lounge booking': ['Passenger Name', 'Passenger Phone', 'Ref NUM', 'Lounge Name', 'Market place', 'Booking Type', 'Date and Time', 'Duration', 'No of Guests', 'Total Amount', 'Payment Status', 'Booking Status'],
    'Bus booking': ['Bus Number', 'Passenger Name', 'Passenger Phone', 'Ref NUM', 'Route', 'Date & Time', 'Bus Type', 'Seat No', 'Total Fare', 'Payment Status', 'Booking Status']
  };

  // Chart data
  busStatusCounts: Record<string, number> = {};
  driverStatusCounts: Record<string, number> = {};
  conductorStatusCounts: Record<string, number> = {};
  loungeRevenueByLounge: { name: string; total: number }[] = [];
  busMonthlyRevenue: number[] = [];
  months: string[] = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
  private colors: string[] = [
    '#0046FF', // Primary blue
    '#10B981', // Emerald green
    '#F59E0B', // Amber
    '#EF4444', // Red
    '#8B5CF6', // Violet
    '#06B6D4', // Cyan
    '#EC4899', // Pink
    '#84CC16', // Lime
    '#F97316', // Orange
    '#6366F1'  // Indigo
  ];

  passengerMonthlyCounts: number[] = [];

  // Multi-year lounge revenue data (mock data)
  selectedYear: string = '2025';
  loungeRevenueYears = [
    { label: 'Year 1', value: '2025' },
    { label: 'Year 2', value: '2023' },
    { label: 'Year 3', value: '2024' }
  ];
  
  loungeRevenueData: { [year: string]: number[] } = {
    '2025': [6, 8, 9, 10, 10.5, 9, 5, 10, 11, 11.4, 11.37, 12.57],
    '2023': [6, 9, 11, 10.5, 11, 11.5, 10, 11, 12, 12.5, 12, 13],
    '2024': [3, 8, 8, 9.5, 11, 10, 5, 10.5, 12, 12.5, 13, 0.8]
  };

  // Experience data (mock)
  driversByExperience = [
    { label: '0-2yrs', count: 70, color: '#3B82F6' },
    { label: '3-5yrs', count: 245, color: '#3B82F6' },
    { label: '6-10yrs', count: 564, color: '#3B82F6' },
    { label: '10+', count: 204, color: '#3B82F6' }
  ];

  conductorsByExperience = [
    { label: '0-2yrs', count: 96, color: '#93C5FD' },
    { label: '3-5yrs', count: 386, color: '#93C5FD' },
    { label: '6-10yrs', count: 701, color: '#93C5FD' },
    { label: '10+', count: 380, color: '#93C5FD' }
  ];

  busByType = [
    { label: 'Normal', count: 825, color: '#93C5FD' },
    { label: 'Semi luxury', count: 550, color: '#93C5FD' },
    { label: 'Luxury', count: 325, color: '#93C5FD' }
  ];

  loungesByCapacity = [
    { label: '20-40', count: 8, color: '#3B82F6' },
    { label: '50-80', count: 10, color: '#3B82F6' },
    { label: '80-300', count: 6, color: '#3B82F6' },
    { label: '200+', count: 12, color: '#3B82F6' }
  ];

  busesByRoute: { label: string; count: number }[] = [];

  constructor(
    private router: Router,
    private busService: BusService,
    private loungeService: LoungeService,
    private driverService: DriverService,
    private conductorService: ConductorService,
    private busBookingService: BusBookingService,
    private loungeBookingService: LoungeBookingService,
    private passengerService: PassengerService,
    public notificationService: NotificationService
  ) {}

  ngOnInit(): void {
    this.refreshPendingNotifications();
    this.pendingNotificationsRefreshId = setInterval(() => {
      this.refreshPendingNotifications();
    }, 30000);

    this.busService.buses$.subscribe(buses => {
      this.numBuses = buses.length;
      this.busStatusCounts = this.countBusStatus(buses);
      this.busesByRoute = this.computeBusesByRoute(buses);
    });

    this.loungeService.lounges$.subscribe(lounges => {
      this.numLounges = lounges.length;
    });

    this.driverService.drivers$.subscribe(drivers => {
      this.numDrivers = drivers.length;
      this.driverStatusCounts = this.countDriverStatus(drivers);
    });

    this.conductorService.conductors$.subscribe(conductors => {
      this.numConductors = conductors.length;
      this.conductorStatusCounts = this.countConductorStatus(conductors);
    });

    this.busBookingService.bookings$.subscribe(bookings => {
      // Calculate total fare for each month for paid bookings (current year)
      const currentYear = new Date().getFullYear();
      const monthlyTotals = Array(12).fill(0);
      bookings.forEach(b => {
        const bookingDate = new Date(b.departure_datetime || b.created_at);
        const bookingYear = bookingDate.getFullYear();
        const month = bookingDate.getMonth();
        
        // Only count paid bookings for the current year
        if (b.payment_status?.toLowerCase() === 'paid' && bookingYear === currentYear) {
          monthlyTotals[month] += b.total_fare;
        }
      });
      this.busMonthlyRevenue = monthlyTotals;
      
      // Update passenger service with bus bookings data for passenger growth chart
      this.passengerService.setBusBookingsData(bookings);
      this.updatePassengerGrowth();
    });

    this.loungeBookingService.bookings$.subscribe(bookings => {
      this.loungeRevenueByLounge = this.computeLoungeRevenue(bookings);
      this.totalLoungeRevenue = bookings.reduce((sum, b) => sum + (b.payment_status === 'paid' ? b.total_amount : 0), 0);
      
      // Update passenger service with lounge bookings data for passenger growth chart
      this.passengerService.setLoungeBookingsData(bookings);
      this.updatePassengerGrowth();
    });

    this.passengerService.passengers$.subscribe(passengers => {
      this.updatePassengerGrowth();
    });
  }

  ngOnDestroy(): void {
    if (this.pendingNotificationsRefreshId) {
      clearInterval(this.pendingNotificationsRefreshId);
      this.pendingNotificationsRefreshId = null;
    }
  }

  private refreshPendingNotifications(): void {
    this.notificationService.loadAllPendingNotifications();
  }

  private updatePassengerGrowth(): void {
    const currentYear = new Date().getFullYear();
    this.passengerMonthlyCounts = this.passengerService.getMonthlyCounts(currentYear);
  }

  private countBusStatus(buses: any[]): Record<string, number> {
    const map: Record<string, number> = { Active: 0, Inactive: 0 };
    buses.forEach(b => {
      const status = typeof b.status === 'string' ? b.status : (b.is_active ? 'Active' : 'Inactive');
      const key = status.toLowerCase() === 'active' ? 'Active' : 'Inactive';
      map[key]++;
    });
    return map;
  }

  private countDriverStatus(drivers: any[]): Record<string, number> {
    const map: Record<string, number> = { Active: 0, Inactive: 0 };
    drivers.forEach(d => {
      const status = typeof d.status === 'string' ? d.status : (d.is_active ? 'Active' : 'Inactive');
      const key = status.toLowerCase() === 'active' ? 'Active' : 'Inactive';
      map[key]++;
    });
    return map;
  }

  private countConductorStatus(conductors: any[]): Record<string, number> {
    const map: Record<string, number> = { Active: 0, 'On Leave': 0, Resigned: 0 };
    conductors.forEach(c => {
      const statusKey = c.status.toLowerCase() === 'active' ? 'Active' : 
                       c.status.toLowerCase() === 'on leave' ? 'On Leave' : 
                       c.status.toLowerCase() === 'resigned' ? 'Resigned' : c.status;
      map[statusKey] = (map[statusKey] || 0) + 1;
    });
    return map;
  }

  private computeLoungeRevenue(bookings: any[]): { name: string; total: number }[] {
    const map: Record<string, number> = {};
    bookings.forEach(b => {
      map[b.lounge_name] = (map[b.lounge_name] || 0) + b.total_amount;
    });
    return Object.entries(map)
      .map(([name, total]) => ({ name, total }))
      .sort((a, b) => b.total - a.total)
      .slice(0, 10); // Limit to top 10 lounges
  }

  private computeBusesByRoute(buses: any[]): { label: string; count: number }[] {
    const map: Record<string, number> = {};
    buses.forEach(b => {
      if (b.custom_route_name) {
        const route = b.custom_route_name.trim();
        map[route] = (map[route] || 0) + 1;
      }
    });
    return Object.entries(map)
      .map(([label, count]) => ({ label, count }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 5); // Top 5 routes
  }

  getBusActivePercentage(): number {
    const total = this.busStatusCounts['Active'] + this.busStatusCounts['Inactive'];
    return total ? (this.busStatusCounts['Active'] / total) * 100 : 0;
  }

  getBusInactivePercentage(): number {
    const total = this.busStatusCounts['Active'] + this.busStatusCounts['Inactive'];
    return total ? (this.busStatusCounts['Inactive'] / total) * 100 : 0;
  }

  getBusPieBackground(): string {
    const active = this.getBusActivePercentage();
    return `conic-gradient(#0046FF 0% ${active}%, #FAA533 ${active}% 100%)`;
  }

  getDriverPieBackground(): string {
    const active = this.getDriverActivePercentage();
    return `conic-gradient(#0046FF 0% ${active}%, #FAA533 ${active}% 100%)`;
  }

  getDriverActivePercentage(): number {
    const total = this.driverStatusCounts['Active'] + this.driverStatusCounts['Inactive'];
    return total ? (this.driverStatusCounts['Active'] / total) * 100 : 0;
  }

  getDriverInactivePercentage(): number {
    const total = this.driverStatusCounts['Active'] + this.driverStatusCounts['Inactive'];
    return total ? (this.driverStatusCounts['Inactive'] / total) * 100 : 0;
  }

  getConductorActivePercentage(): number {
    const total = Object.values(this.conductorStatusCounts).reduce((a, b) => a + b, 0);
    return total ? (this.conductorStatusCounts['Active'] / total) * 100 : 0;
  }

  getConductorOnLeavePercentage(): number {
    const total = Object.values(this.conductorStatusCounts).reduce((a, b) => a + b, 0);
    return total ? (this.conductorStatusCounts['On Leave'] / total) * 100 : 0;
  }

  getConductorResignedPercentage(): number {
    const total = Object.values(this.conductorStatusCounts).reduce((a, b) => a + b, 0);
    return total ? (this.conductorStatusCounts['Resigned'] / total) * 100 : 0;
  }

  getConductorPieBackground(): string {
    const active = this.getConductorActivePercentage();
    const onLeave = this.getConductorOnLeavePercentage();
    const resigned = this.getConductorResignedPercentage();
    const activeEnd = active;
    const onLeaveEnd = active + onLeave;
    return `conic-gradient(#0046FF 0% ${activeEnd}%, #FAA533 ${activeEnd}% ${onLeaveEnd}%, #9ca3af ${onLeaveEnd}% 100%)`;
  }

  getRevenuePoints(): string {
    const max = Math.max(...this.busMonthlyRevenue, 1);
    return this.busMonthlyRevenue.map((v, i) => `${i * 50 + 50},${260 - (v / max) * 200}`).join(' ');
  }

  getYAxisLabels(): { value: string, y: number }[] {
    const max = Math.max(...this.busMonthlyRevenue, 1);
    const steps = [0, 0.25, 0.5, 0.75, 1];
    return steps.map(f => {
      const val = f * max;
      const y = 260 - (val / max) * 200;
      return { value: Math.round(val).toString(), y };
    });
  }

  getLoungeRevenueBarHeight(total: number): number {
    const max = Math.max(...this.loungeRevenueByLounge.map(i => i.total), 1);
    return (total / max) * 100;
  }

  getColorForLounge(name: string): string {
    const index = this.loungeRevenueByLounge.findIndex(item => item.name === name);
    return this.colors[index % this.colors.length];
  }

  getPassengerBarHeight(count: number): number {
    const max = Math.max(...this.passengerMonthlyCounts, 1);
    return (count / max) * 100;
  }

  getColorForPassenger(index: number): string {
    return '#0046FF';
  }

  toggleNotificationPanel(): void {
    this.showNotificationPanel = !this.showNotificationPanel;
  }

  closeNotificationPanel(): void {
    this.showNotificationPanel = false;
  }

  showProfileMenu = false;

  toggleProfileMenu() {
    this.showProfileMenu = !this.showProfileMenu;
  }

  logout() {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('admin_user');
    this.router.navigate(['/login']);
  }

  // Search functionality methods
  selectSearchType(type: string) {
    this.selectedSearchType = type;
    this.selectedAttributes = [];
    this.searchValues = {};
    this.showSearchResults = false;
  }

  toggleAttribute(attribute: string) {
    const index = this.selectedAttributes.indexOf(attribute);
    if (index > -1) {
      this.selectedAttributes.splice(index, 1);
      delete this.searchValues[attribute];
    } else {
      this.selectedAttributes.push(attribute);
      this.searchValues[attribute] = '';
    }
  }

  isAttributeSelected(attribute: string): boolean {
    return this.selectedAttributes.includes(attribute);
  }

  performSearch() {
    if (this.selectedAttributes.length === 0) {
      alert('Please select at least one search attribute');
      return;
    }

    // Check if at least one search value is provided
    const hasValue = this.selectedAttributes.some(attr => this.searchValues[attr]?.trim());
    if (!hasValue) {
      alert('Please enter at least one search value');
      return;
    }

    // Navigate to search results page with query params
    this.router.navigate(['/search-results'], {
      queryParams: {
        type: this.selectedSearchType,
        criteria: JSON.stringify(this.searchValues)
      }
    });
  }

  private searchBuses() {
    const buses = this.busService.buses; // Get all buses
    this.searchResults = buses.filter(bus => {
      return this.selectedAttributes.every(attr => {
        const searchValue = this.searchValues[attr]?.toLowerCase().trim();
        if (!searchValue) return true;

        switch (attr) {
          case 'Company':
            return bus.company_name?.toLowerCase().includes(searchValue);
          case 'Route':
            return bus.custom_route_name?.toLowerCase().includes(searchValue);
          case 'Permit Num':
            return bus.permit_number?.toLowerCase().includes(searchValue);
          case 'Register Num':
            return bus.license_plate?.toLowerCase().includes(searchValue);
          case 'Owner Verification':
            return bus.owner_verification_status?.toLowerCase().includes(searchValue);
          case 'Permit Verify':
            return bus.verification_status?.toLowerCase().includes(searchValue);
          case 'Contact':
            return bus.business_phone?.toLowerCase().includes(searchValue);
          case 'No of Seat':
            return bus.total_seats?.toString().includes(searchValue);
          case 'Approved fare':
            return bus.fare_per_seat?.toString().includes(searchValue);
          case 'Type':
            return bus.bus_type?.toLowerCase().includes(searchValue);
          case 'Status':
            return bus.status?.toLowerCase().includes(searchValue);
          default:
            return true;
        }
      });
    });
  }

  private searchLounges() {
    const lounges = this.loungeService.lounges; // Get all lounges
    this.searchResults = lounges.filter(lounge => {
      return this.selectedAttributes.every(attr => {
        const searchValue = this.searchValues[attr]?.toLowerCase();
        if (!searchValue) return true;

        switch (attr) {
          case 'Lounge Name':
            return lounge.lounge_name?.toLowerCase().includes(searchValue);
          case 'Owner':
            return lounge.lounge_owner?.toLowerCase().includes(searchValue);
          case 'Contact':
            return lounge.lounge_contact?.toLowerCase().includes(searchValue);
          case 'Address':
            return lounge.address?.toLowerCase().includes(searchValue);
          case 'Price per hour':
            return lounge.price_per_hour?.toString().includes(searchValue);
          case 'Capacity':
            return lounge.capacity?.toString().includes(searchValue);
          case 'Operation':
            const status = typeof lounge.operational === 'string' ? lounge.operational : (lounge.operational ? 'open' : 'closed');
            return status.toLowerCase().includes(searchValue);
          case 'Verification':
            return lounge.verification?.toLowerCase().includes(searchValue);
          default:
            return true;
        }
      });
    });
  }

  private searchDrivers() {
    const drivers = this.driverService.drivers; // Get all drivers
    this.searchResults = drivers.filter(driver => {
      return this.selectedAttributes.every(attr => {
        const searchValue = this.searchValues[attr]?.toLowerCase().trim();
        if (!searchValue) return true;

        switch (attr) {
          case 'Name':
            return driver.name?.toLowerCase().includes(searchValue);
          case 'Contact':
            return driver.contact_number?.toLowerCase().includes(searchValue);
          case 'License Num':
            return driver.license_number?.toLowerCase().includes(searchValue);
          case 'License Expire date':
            return driver.license_expiry_date?.toLowerCase().includes(searchValue);
          case 'Experience':
            return driver.experience_years?.toString().includes(searchValue);
          case 'Hire date':
            return driver.hire_date?.toLowerCase().includes(searchValue);
          case 'Verification':
            return driver.verification_status?.toLowerCase().includes(searchValue);
          case 'Status':
            return driver.status?.toLowerCase() === searchValue;
          default:
            return true;
        }
      });
    });
  }

  private searchConductors() {
    const conductors = this.conductorService.conductors; // Get all conductors
    this.searchResults = conductors.filter(conductor => {
      return this.selectedAttributes.every(attr => {
        const searchValue = this.searchValues[attr]?.toLowerCase().trim();
        if (!searchValue) return true;

        switch (attr) {
          case 'Name':
            return conductor.name?.toLowerCase().includes(searchValue);
          case 'Contact':
            return conductor.contact_number?.toLowerCase().includes(searchValue);
          case 'License Num':
            return conductor.license_number?.toLowerCase().includes(searchValue);
          case 'Experience':
            return conductor.experience_years?.toString().includes(searchValue);
          case 'Hire date':
            return conductor.hire_date?.toLowerCase().includes(searchValue);
          case 'Verification':
            return conductor.verification_status?.toLowerCase().includes(searchValue);
          case 'Status':
            return conductor.status?.toLowerCase() === searchValue;
          default:
            return true;
        }
      });
    });
  }

  private searchBusBookings() {
    const bookings = this.busBookingService.bookings; // Get all bus bookings
    this.searchResults = bookings.filter(booking => {
      return this.selectedAttributes.every(attr => {
        const searchValue = this.searchValues[attr]?.toLowerCase().trim();
        if (!searchValue) return true;

        switch (attr) {
          case 'Bus Number':
            return booking.bus_number?.toLowerCase().includes(searchValue);
          case 'Passenger Name':
            return booking.passenger_name?.toLowerCase().includes(searchValue);
          case 'Passenger Phone':
            return booking.passenger_phone?.toLowerCase().includes(searchValue);
          case 'Ref NUM':
            return booking.booking_reference?.toLowerCase().includes(searchValue);
          case 'Route':
            return booking.route?.toLowerCase().includes(searchValue);
          case 'Date & Time':
            return booking.departure_datetime?.toLowerCase().includes(searchValue);
          case 'Bus Type':
            return booking.bus_type?.toLowerCase().includes(searchValue);
          case 'Seat No':
            return booking.seat_number?.toLowerCase().includes(searchValue);
          case 'Total Fare':
            return booking.total_fare?.toString().includes(searchValue);
          case 'Payment Status':
            return booking.payment_status?.toLowerCase().includes(searchValue);
          case 'Booking Status':
            return booking.booking_status?.toLowerCase().includes(searchValue);
          default:
            return true;
        }
      });
    });
  }

  private searchLoungeBookings() {
    const bookings = this.loungeBookingService.bookings; // Get all lounge bookings
    this.searchResults = bookings.filter(booking => {
      return this.selectedAttributes.every(attr => {
        const searchValue = this.searchValues[attr]?.toLowerCase();
        if (!searchValue) return true;

        switch (attr) {
          case 'Passenger Name':
            return booking.passenger_name?.toLowerCase().includes(searchValue);
          case 'Passenger Phone':
            return booking.passenger_phone?.toLowerCase().includes(searchValue);
          case 'Ref NUM':
            return booking.booking_reference?.toLowerCase().includes(searchValue);
          case 'Lounge Name':
            return booking.lounge_name?.toLowerCase().includes(searchValue);
          case 'Market place':
            return booking.product_name?.toLowerCase().includes(searchValue);
          case 'Booking Type':
            return booking.booking_type?.toLowerCase().includes(searchValue);
          case 'Date and Time':
            return booking.scheduled_arrival?.toLowerCase().includes(searchValue);
          case 'Duration':
            return booking.pricing_type?.toLowerCase().includes(searchValue);
          case 'No of Guests':
            return booking.number_of_guests?.toString().includes(searchValue);
          case 'Total Amount':
            return booking.total_amount?.toString().includes(searchValue);
          case 'Payment Status':
            return booking.payment_status?.toLowerCase().includes(searchValue);
          case 'Booking Status':
            return booking.status?.toLowerCase().includes(searchValue);
          default:
            return true;
        }
      });
    });
  }

  getSearchResultColumns(): string[] {
    switch (this.selectedSearchType) {
      case 'Bus':
        return ['BusID', 'Company', 'Phone Num', 'NIC Num', 'Email', 'Permit Num', 'Registered Num', 'Route', 'Route(via)', 'Approved Fare', 'Validity period', 'Type', 'Seats', 'Status', 'Action'];
      case 'Lounge':
        return ['Lounge ID', 'Lounge Name', 'Owner', 'Capacity', 'Price/Hour', 'Facilities', 'Marketplace', 'Verification', 'Status', 'Action'];
      case 'Driver':
        return ['Driver ID', 'Name', 'Contact', 'License Number', 'License Expiry', 'Experience Yrs', 'Verification', 'Status', 'Action'];
      case 'Conductor':
        return ['Conductor ID', 'Name', 'Contact', 'License Number', 'License Expiry', 'Experience Yrs', 'Verification', 'Status', 'Action'];
      case 'Bus booking':
        return ['Booking ID', 'Passenger ID', 'Bus ID', 'Departure', 'Seats', 'Total Fare', 'Payment Status', 'Booking Status', 'Action'];
      case 'Lounge booking':
        return ['L_Booking ID', 'Passenger Name', 'Passenger Phone', 'Ref NUM', 'Lounge Name', 'Market place', 'Booking Type', 'Date and Time', 'Duration', 'No of Guests', 'Total Amount', 'Payment Status', 'Booking Status'];
      default:
        return [];
    }
  }

  clearSearch() {
    this.selectedAttributes = [];
    this.searchValues = {};
    this.searchResults = [];
    this.showSearchResults = false;
  }

  // Multi-year lounge revenue chart methods
  selectLoungeYear(year: string) {
    this.selectedYear = year;
  }

  getLoungeRevenuePoints(): string {
    const data = this.loungeRevenueData[this.selectedYear] || [];
    const max = 20; // Fixed scale from 0 to 20
    return data.map((v, i) => `${i * 50 + 50},${260 - (v / max) * 200}`).join(' ');
  }

  getLoungeYAxisLabels(): { value: string, y: number }[] {
    const max = 20;
    const steps = [0, 0.25, 0.5, 0.75, 1];
    return steps.map(f => {
      const val = f * max;
      const y = 260 - (val / max) * 200;
      return { value: val.toString(), y };
    });
  }

  getLineColor(year: string): string {
    const colors: { [key: string]: string } = {
      '2025': '#74C0FC',  // Blue
      '2023': '#FFB27D',  // Orange
      '2024': '#FFE66D'   // Yellow
    };
    return colors[year] || '#74C0FC';
  }

  getLoungeRevenuePointsForYear(year: string): string {
    const data = this.loungeRevenueData[year] || [];
    const max = 20;
    return data.map((v, i) => `${i * 50 + 50},${210 - (v / max) * 160}`).join(' ');
  }

  getLoungeRevenueDotY(year: string, monthIndex: number): number {
    const data = this.loungeRevenueData[year] || [];
    const value = data[monthIndex] || 0;
    const max = 20;
    return 220 - (value / max) * 160;
  }

  // Horizontal bar chart methods
  getHorizontalBarWidth(count: number, data: any[]): number {
    const max = Math.max(...data.map(d => d.count), 1);
    return (count / max) * 100;
  }
}
