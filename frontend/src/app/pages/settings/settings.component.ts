import { Component, OnInit, Inject, PLATFORM_ID } from '@angular/core';
import { CommonModule, isPlatformBrowser } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { InputTextModule } from 'primeng/inputtext';
import { ButtonModule } from 'primeng/button';
import { AvatarModule } from 'primeng/avatar';
import { TableModule } from 'primeng/table';
import { TagModule } from 'primeng/tag';
import { CheckboxModule } from 'primeng/checkbox';
import { ToggleSwitchModule } from 'primeng/toggleswitch';
import { NavbarComponent } from '../../shared/components/navbar/navbar.component';
import { NotificationPanelComponent } from '../../shared/components/notification-panel/notification-panel.component';
import { NotificationService } from '../../core/services/notification.service';
import { AdminAuthService } from '../../core/services/admin-auth.service';
import { UserManagementService } from '../../core/services/user-management.service';

@Component({
  selector: 'app-settings',
  standalone: true,
  imports: [
    CommonModule, 
    FormsModule, 
    InputTextModule, 
    ButtonModule, 
    AvatarModule,
    TableModule,
    TagModule,
    CheckboxModule,
    ToggleSwitchModule,
    NavbarComponent,
    NotificationPanelComponent,
    RouterModule
  ],
  templateUrl: './settings.component.html',
  styleUrls: ['./settings.component.scss']
})
export class SettingsComponent implements OnInit {
  activeSection: string = 'Profile Setting';
  showNotificationPanel = false;
  private isBrowser: boolean;
  private originalTheme: string = 'system';
  isSuperAdmin: boolean = false;
  filteredMenuItems: string[] = [];
  
  userProfile = {
    fullName: 'Dinesh Priyash',
    userName: 'Din@Sh',
    accessLevel: 'Super Admin',
    dateJoined: '12/10/2025',
    contactNumber: '070 2345678',
    emergencyContact: '078 2345678',
    workEmail: 'Dinesh@Myprogmail.com'
  };

  userViewMode: 'list' | 'add' | 'edit' | 'view' = 'list';
  
  newUser = {
    fullName: '',
    userName: '',
    email: '',
    contactNumber: '',
    emergencyContact: '',
    role: 'admin' as 'admin' | 'supervisor' | 'super_admin',
    appScope: 'bus' as '' | 'bus' | 'driver' | 'lounges' | 'passenger',
    supervisorId: '',
    password: '',
    confirmPassword: ''
  };

  selectedUser: any = {};
  supervisorOptions: Array<{ label: string; value: string; appScope: string }> = [];
  availableSupervisors: Array<{ label: string; value: string; appScope: string }> = [];

  readonly roleOptions = [
    { label: 'Admin', value: 'admin' },
    { label: 'Supervisor', value: 'supervisor' },
    { label: 'Super Admin', value: 'super_admin' }
  ];

  readonly appScopeOptions = [
    { label: 'Bus', value: 'bus' },
    { label: 'Driver', value: 'driver' },
    { label: 'Lounges', value: 'lounges' },
    { label: 'Passenger', value: 'passenger' }
  ];

  permissionGroups = [
    {
      name: 'Bus Management',
      selected: false,
      permissions: [
        { name: 'View Buses', selected: false },
        { name: 'Add Bus', selected: false },
        { name: 'Edit Bus', selected: false },
        { name: 'Approve/Reject New Buses', selected: false },
        { name: 'View & Download reports', selected: false }
      ]
    },
    {
      name: 'Driver & Conductor Management',
      selected: false,
      permissions: [
        { name: 'View Drivers & Conductors', selected: false },
        { name: 'Add Drivers & Conductors', selected: false },
        { name: 'Edit Drivers & Conductors', selected: false },
        { name: 'Approve/Reject', selected: false },
        { name: 'View & Download reports', selected: false }
      ]
    },
    {
      name: 'Lounge Management',
      selected: false,
      permissions: [
        { name: 'View Lounges', selected: false },
        { name: 'Add Lounges', selected: false },
        { name: 'Edit Lounges', selected: false },
        { name: 'Approve/Reject Lounges', selected: false },
        { name: 'View & Download reports', selected: false }
      ]
    },
    {
      name: 'Bus & Lounge Bookings',
      selected: false,
      permissions: [
        { name: 'View Bookings', selected: false },
        { name: 'Add Bookings', selected: false },
        { name: 'Edit Bookings When user want', selected: false },
        { name: 'View & Download reports', selected: false }
      ]
    },
    {
      name: 'Complaint',
      selected: false,
      permissions: [
        { name: 'View Complaints', selected: false },
        { name: 'Manage Complaints', selected: false },
        { name: 'Send Solutions', selected: false },
        { name: 'Add Remarks', selected: false },
        { name: 'View & Download reports', selected: false }
      ]
    },
    {
      name: 'Feedbacks',
      selected: false,
      permissions: [
        { name: 'View Feedbacks', selected: false },
        { name: 'Send Replys', selected: false }
      ]
    },
    {
      name: 'Users & Roles',
      selected: false,
      permissions: [
        { name: 'View Users', selected: false },
        { name: 'Add Users', selected: false },
        { name: 'Edit Users', selected: false },
        { name: 'Assign roles', selected: false }
      ]
    },
    {
      name: 'System Configuration',
      selected: false,
      permissions: [
        { name: 'View Setting', selected: false },
        { name: 'Manage Setting', selected: false }
      ]
    },
    {
      name: 'Notifications',
      selected: false,
      permissions: [
        { name: 'View Notifications', selected: false },
        { name: 'Manage Notifications', selected: false },
        { name: 'Send Notifications', selected: false },
        { name: 'Notification Settings', selected: false }
      ]
    }
  ];

  users: any[] = [];

  // All possible menu items
  private allMenuItems = [
    'Profile Setting',
    'Users & Roles',
    'Notification Settings',
    'System Appearance',
    'Security & Privacy'
  ];

  // Menu items accessible to both admin and super_admin
  private commonMenuItems = [
    'Profile Setting',
    'System Appearance'
  ];

  // Menu items only for super_admin
  private superAdminOnlyItems = [
    'Users & Roles'
  ];

  notificationSettings = {
    busNotification: true,
    driverNotification: true,
    conductorNotification: false,
    loungesNotification: true,
    loungeBookingNotification: true,
    busBookingNotification: true,
    complaintsNotification: true,
    complaints: true,
    
    quietHours: {
      enabled: true,
      duration: '1 hour'
    },
    desktopNotifications: true,
    unreadBadge: true,
    notificationSounds: true,
    emailAlerts: true,
    autoMarkRead: true
  };

  systemPreferences = {
    theme: 'light',
    language: 'English (United States)',
    timezone: '(UTC-08:00) Pacific Time (US & Canada)',
    dateFormat: 'MM/DD/YYYY'
  };

  securitySettings = {
    currentPassword: '',
    newPassword: '',
    confirmPassword: '',
    allowNewDeviceLogin: true,
    lastLogin: '2025-12-10 10:45 AM',
    loggedInIp: '192.168.1.24',
    requirePasswordForSensitive: true,
    loginAlerts: true,
    failedLoginProtection: true
  };



  constructor(
    private router: Router, 
    @Inject(PLATFORM_ID) private platformId: Object, 
    public notificationService: NotificationService,
    private authService: AdminAuthService,
    private userManagementService: UserManagementService
  ) {
    this.isBrowser = isPlatformBrowser(this.platformId);
  }

  ngOnInit() {
    // Check user role and filter menu items
    this.isSuperAdmin = this.authService.isSuperAdmin();
    this.filterMenuItems();

    // Load current user profile
    const currentUser = this.authService.getCurrentAdmin();
    if (currentUser) {
      const normalizedRole = (currentUser.role || '').toLowerCase().trim().replace(/[-\s]+/g, '_');
      this.userProfile = {
        fullName: currentUser.full_name,
        userName: currentUser.email.split('@')[0],
        accessLevel: normalizedRole === 'super_admin' ? 'Super Admin' : 'Admin',
        dateJoined: new Date(currentUser.created_at).toLocaleDateString(),
        contactNumber: '', // Not available in current user model
        emergencyContact: '',
        workEmail: currentUser.email
      };
    }

    // Load users if super admin
    if (this.isSuperAdmin) {
      this.loadUsers();
    }

    if (this.isBrowser) {
      const savedTheme = localStorage.getItem('theme');
      if (savedTheme) {
        this.systemPreferences.theme = savedTheme;
        this.originalTheme = savedTheme;
      } else {
        this.systemPreferences.theme = 'system';
        this.originalTheme = 'system';
      }
    }
  }

  private filterMenuItems() {
    if (this.isSuperAdmin) {
      // Super admin sees common items + super admin only items
      this.filteredMenuItems = [...this.commonMenuItems, ...this.superAdminOnlyItems];
    } else {
      // Regular admin sees only common items
      this.filteredMenuItems = [...this.commonMenuItems];
    }
  }

  private loadUsers() {
    this.userManagementService.getAllUsers().subscribe({
      next: (response) => {
        this.users = response.users.map(user => ({
          id: user.id,
          fullName: user.full_name,
          userName: user.email.split('@')[0],
          contact: '', // Not available in current model
          email: user.email,
          lastActivity: '', // Not tracked yet
          role: this.toRoleLabel(user.role),
          roleValue: user.role,
          appScope: user.app_scope ?? '',
          supervisorId: user.supervisor_id ?? '',
          permissionsArray: user.permissions ?? [],
          permissions: user.role === 'super_admin'
            ? 'Full Access'
            : (user.permissions && user.permissions.length > 0 ? user.permissions.join(', ') : 'Restricted'),
          status: user.is_active ? 'Active' : 'Inactive'
        }));

        this.rebuildSupervisorOptionsFromUsers();
      },
      error: (error) => {
        console.error('Error loading users:', error);
      }
    });
  }

  private rebuildSupervisorOptionsFromUsers() {
    this.supervisorOptions = this.users
      .filter((user: any) => user.roleValue === 'supervisor' || user.roleValue === 'super_admin')
      .map((user: any) => ({
        label: `${user.fullName} (${this.toRoleLabel(user.roleValue)})`,
        value: user.id,
        appScope: user.appScope ?? ''
      }));

    this.updateAvailableSupervisors();
  }

  private toRoleLabel(role: string): string {
    if (role === 'super_admin') return 'Super Admin';
    if (role === 'supervisor') return 'Supervisor';
    return 'Admin';
  }

  private toRoleValue(label: string): 'admin' | 'supervisor' | 'super_admin' {
    const normalized = label.toLowerCase();
    if (normalized.includes('super admin')) return 'super_admin';
    if (normalized.includes('supervisor')) return 'supervisor';
    return 'admin';
  }

  onNewUserRoleChange() {
    if (this.newUser.role === 'super_admin') {
      this.newUser.appScope = '';
    }

    if (this.newUser.role !== 'admin') {
      this.newUser.supervisorId = '';
    }

    this.updateAvailableSupervisors();
  }

  onNewUserAppScopeChange() {
    this.newUser.supervisorId = '';
    this.updateAvailableSupervisors();
  }

  private updateAvailableSupervisors() {
    if (this.newUser.role !== 'admin') {
      this.availableSupervisors = [];
      return;
    }

    this.availableSupervisors = this.supervisorOptions.filter(s => {
      const isSuperAdmin = s.label.toLowerCase().includes('super admin');
      return isSuperAdmin || s.appScope === this.newUser.appScope;
    });
  }

  selectTheme(theme: string) {
    this.systemPreferences.theme = theme;
    if (this.isBrowser) {
      if (theme === 'system') {
        this.applySystemTheme();
      } else {
        this.applyTheme(theme);
      }
    }
  }

  toggleGroupPermissions(group: any) {
    if (group.permissions) {
      group.permissions.forEach((permission: any) => {
        permission.selected = group.selected;
      });
    }
  }

  checkGroupPermissions(group: any) {
    if (group.permissions) {
      if (group.name === 'System Configuration' || group.name === 'Notifications') {
        // For these groups, parent is checked if ANY child is checked
        group.selected = group.permissions.some((permission: any) => permission.selected);
      } else {
        // For other groups, parent is checked only if ALL children are checked
        const allSelected = group.permissions.every((permission: any) => permission.selected);
        group.selected = allSelected;
      }
    }
  }

  setActiveSection(section: string) {
    this.activeSection = section;
    this.userViewMode = 'list';

    if (section === 'Users & Roles' && this.isSuperAdmin) {
      this.loadUsers();
    }
  }

  resetPermissions() {
    this.permissionGroups.forEach(group => {
      group.selected = false;
      group.permissions.forEach(p => p.selected = false);
    });
  }

  updatePermissionsBasedOnUser(user: any) {
    this.resetPermissions();
    
    if (user.permissions === 'Full Access') {
      this.permissionGroups.forEach(group => {
        group.selected = true;
        group.permissions.forEach(p => p.selected = true);
      });
    } else {
      const perms = user.permissions.toLowerCase();
      
      this.permissionGroups.forEach(group => {
        let match = false;
        
        // Map permission string keywords to groups
        if (group.name === 'Bus Management' && (perms.includes('bus'))) match = true;
        else if (group.name === 'Driver & Conductor Management' && (perms.includes('driver') || perms.includes('conductor'))) match = true;
        else if (group.name === 'Lounge Management' && (perms.includes('lounge') || perms.includes('loung'))) match = true;
        else if (group.name === 'Bus & Lounge Bookings' && (perms.includes('booking'))) match = true;
        else if (group.name === 'Complaint' && (perms.includes('complaint'))) match = true;
        else if (group.name === 'Feedbacks' && (perms.includes('feedback'))) match = true;
        else if (group.name === 'Users & Roles' && (perms.includes('user') || perms.includes('role'))) match = true;
        else if (group.name === 'System Configuration' && (perms.includes('setting') || perms.includes('config'))) match = true;
        else if (group.name === 'Notifications' && (perms.includes('notification'))) match = true;

        if (match) {
          group.selected = true;
          group.permissions.forEach(p => p.selected = true);
        }
      });
    }
  }

  showAddUser() {
    this.userViewMode = 'add';
    this.resetNewUser();
    this.onNewUserRoleChange();
    this.resetPermissions();
  }

  editUser(user: any) {
    this.userViewMode = 'edit';
    this.selectedUser = { 
      ...user, 
      contactNumber: user.contact,
      isActive: user.status.toLowerCase() === 'active',
      roleValue: user.roleValue ?? this.toRoleValue(user.role),
      appScope: user.appScope ?? '',
      supervisorId: user.supervisorId ?? ''
    }; 
    this.applyPermissionCodesToUI(user.permissionsArray ?? []);
  }

  viewUser(user: any) {
    this.userViewMode = 'view';
    this.selectedUser = { 
      ...user, 
      contactNumber: user.contact,
      isActive: user.status.toLowerCase() === 'active',
      roleValue: user.roleValue ?? this.toRoleValue(user.role),
      appScope: user.appScope ?? '',
      supervisorId: user.supervisorId ?? ''
    };
    this.applyPermissionCodesToUI(user.permissionsArray ?? []);
  }

  cancelAddUser() {
    this.userViewMode = 'list';
    this.selectedUser = {};
  }

  resetNewUser() {
    this.newUser = {
      fullName: '',
      userName: '',
      email: '',
      contactNumber: '',
      emergencyContact: '',
      role: 'admin',
      appScope: 'bus',
      supervisorId: '',
      password: '',
      confirmPassword: ''
    };
  }

  saveNewUser() {
    if (!this.isSuperAdmin) {
      alert('Only super admins can create users.');
      return;
    }

    const email = (this.newUser.email || '').trim();
    const fullName = (this.newUser.fullName || '').trim();
    const password = this.newUser.password || '';
    const confirmPassword = this.newUser.confirmPassword || '';

    if (!fullName || !email || !password || !confirmPassword) {
      alert('Please fill in full name, email, password, and confirm password.');
      return;
    }

    // Validate password match
    if (password !== confirmPassword) {
      alert('Passwords do not match.');
      return;
    }

    if (password.length < 8) {
      alert('Password must be at least 8 characters long.');
      return;
    }

    if (this.newUser.role === 'admin' && !this.newUser.supervisorId) {
      alert('Please select a supervisor for admin users.');
      return;
    }

    if (this.newUser.role !== 'super_admin' && !this.newUser.appScope) {
      alert('Please select an app scope.');
      return;
    }

    const selectedPermissions = this.getSelectedFunctionPermissions(this.newUser.role);

    if (this.newUser.role !== 'super_admin' && selectedPermissions.length === 0) {
      alert('Please select at least one function permission.');
      return;
    }

    // Call API to create user
    this.userManagementService.createUser({
      email,
      full_name: fullName,
      password,
      role: this.newUser.role,
      app_scope: this.newUser.role === 'super_admin' ? undefined : (this.newUser.appScope || undefined),
      supervisor_id: this.newUser.role === 'admin' ? (this.newUser.supervisorId || undefined) : undefined,
      permissions: selectedPermissions
    }).subscribe({
      next: (createdUser) => {
        // Add to local list
        const newUserEntry = {
          id: createdUser.id,
          fullName: createdUser.full_name,
          userName: createdUser.email.split('@')[0],
          contact: this.newUser.contactNumber,
          email: createdUser.email,
          lastActivity: 'Just now',
          role: this.toRoleLabel(createdUser.role),
          roleValue: createdUser.role,
          appScope: createdUser.app_scope ?? '',
          supervisorId: createdUser.supervisor_id ?? '',
          permissionsArray: createdUser.permissions ?? selectedPermissions,
          permissions: createdUser.role === 'super_admin'
            ? 'Full Access'
            : ((createdUser.permissions ?? selectedPermissions).join(', ') || 'Restricted'),
          status: createdUser.is_active ? 'Active' : 'Inactive'
        };

        this.users = [...this.users, newUserEntry];
        this.rebuildSupervisorOptionsFromUsers();
        this.userViewMode = 'list';
        this.resetNewUser();
        this.resetPermissions();
      },
      error: (error) => {
        console.error('Error creating user:', error);
        alert('Failed to create user: ' + (error.error?.error || 'Unknown error'));
      }
    });
  }

  updateUser() {
    const index = this.users.findIndex(u => u.id === this.selectedUser.id);
    if (index !== -1) {
      const selectedPermissions = this.getSelectedFunctionPermissions(this.selectedUser.roleValue);

      if (this.selectedUser.roleValue === 'admin' && !this.selectedUser.supervisorId) {
        alert('Please select a supervisor for admin users.');
        return;
      }

      if (this.selectedUser.roleValue !== 'super_admin' && selectedPermissions.length === 0) {
        alert('Please select at least one function permission.');
        return;
      }

      // Call API to update user
      this.userManagementService.updateUser(this.selectedUser.id, {
        full_name: this.selectedUser.fullName,
        is_active: this.selectedUser.isActive,
        role: this.selectedUser.roleValue,
        app_scope: this.selectedUser.roleValue === 'super_admin' ? undefined : (this.selectedUser.appScope || undefined),
        supervisor_id: this.selectedUser.roleValue === 'admin' ? (this.selectedUser.supervisorId || undefined) : undefined,
        permissions: selectedPermissions
      }).subscribe({
        next: (updatedUser) => {
          // Update local list
          const updatedUserEntry = {
            ...this.users[index],
            fullName: updatedUser.full_name,
            email: updatedUser.email,
            role: this.toRoleLabel(updatedUser.role),
            roleValue: updatedUser.role,
            appScope: updatedUser.app_scope ?? '',
            supervisorId: updatedUser.supervisor_id ?? '',
            permissionsArray: updatedUser.permissions ?? selectedPermissions,
            permissions: updatedUser.role === 'super_admin'
              ? 'Full Access'
              : ((updatedUser.permissions ?? selectedPermissions).join(', ') || 'Restricted'),
            status: updatedUser.is_active ? 'Active' : 'Inactive'
          };

          const updatedUsers = [...this.users];
          updatedUsers[index] = updatedUserEntry;
          this.users = updatedUsers;
          this.rebuildSupervisorOptionsFromUsers();
          
          this.userViewMode = 'list';
          this.selectedUser = {};
          this.resetPermissions();
        },
        error: (error) => {
          console.error('Error updating user:', error);
          alert('Failed to update user: ' + (error.error?.error || 'Unknown error'));
        }
      });
    }
  }

  private calculatePermissionsString(): string {
    const allGroupsSelected = this.permissionGroups.every(g => g.selected && g.permissions.every(p => p.selected));
    
    if (allGroupsSelected) {
      return 'Full Access';
    } else {
      const selectedGroupNames = this.permissionGroups
        .filter(g => g.selected || g.permissions.some(p => p.selected))
        .map(g => {
            // Simplified mapping for display
            if (g.name.includes('Bus Management')) return 'Manage buses';
            if (g.name.includes('Driver')) return 'drivers';
            if (g.name.includes('Lounge')) return 'lounges';
            if (g.name.includes('Booking')) return 'bookings';
            return g.name;
        });
      
      if (selectedGroupNames.length > 0) {
          let permissionsStr = selectedGroupNames.join(', ');
          if (permissionsStr.length > 30) {
              permissionsStr = permissionsStr.substring(0, 30) + '...';
          }
          return permissionsStr;
      } else {
          return 'Restricted';
      }
    }
  }

  private getSelectedFunctionPermissions(role: 'admin' | 'supervisor' | 'super_admin'): string[] {
    if (role === 'super_admin') {
      return ['*'];
    }

    const permissions = new Set<string>();
    const hasAnySelection = this.permissionGroups.some(g => g.permissions.some(p => p.selected));
    if (!hasAnySelection) {
      return [];
    }

    const selectedByGroup = (groupName: string): boolean => {
      const group = this.permissionGroups.find(g => g.name === groupName);
      return !!group && group.permissions.some(p => p.selected);
    };

    if (selectedByGroup('Bus Management')) {
      permissions.add('buses.read');
      permissions.add('buses.write');
      permissions.add('buses.verify');
      permissions.add('bus_owners.read');
      permissions.add('bus_owners.write');
      permissions.add('bus_owners.verify');
    }

    if (selectedByGroup('Driver & Conductor Management')) {
      permissions.add('drivers.read');
      permissions.add('drivers.write');
      permissions.add('drivers.verify');
      permissions.add('conductors.read');
      permissions.add('conductors.write');
      permissions.add('conductors.verify');
    }

    if (selectedByGroup('Lounge Management')) {
      permissions.add('lounges.read');
      permissions.add('lounges.write');
      permissions.add('lounges.verify');
      permissions.add('lounge_owners.read');
      permissions.add('lounge_owners.verify');
    }

    if (selectedByGroup('Bus & Lounge Bookings')) {
      permissions.add('bookings.read');
      permissions.add('bookings.write');
      permissions.add('lounge_bookings.read');
      permissions.add('lounge_bookings.write');
    }

    if (selectedByGroup('Complaint')) {
      permissions.add('complaints.read');
      permissions.add('complaints.update');
      permissions.add('complaints.escalate');
      permissions.add('escalation.read');
      permissions.add('escalation.manage');
    }

    if (selectedByGroup('Users & Roles')) {
      permissions.add('users.manage');
      permissions.add('permissions.manage');
    }

    return Array.from(permissions).sort();
  }

  private applyPermissionCodesToUI(permissionCodes: string[]) {
    this.resetPermissions();
    if (!permissionCodes || permissionCodes.length === 0) {
      return;
    }

    if (permissionCodes.includes('*')) {
      this.permissionGroups.forEach(group => {
        group.selected = true;
        group.permissions.forEach(p => p.selected = true);
      });
      return;
    }

    const markGroup = (groupName: string, shouldMark: boolean) => {
      if (!shouldMark) {
        return;
      }
      const group = this.permissionGroups.find(g => g.name === groupName);
      if (!group) {
        return;
      }
      group.selected = true;
      group.permissions.forEach(p => p.selected = true);
    };

    const hasAny = (...codes: string[]) => codes.some(code => permissionCodes.includes(code));

    markGroup('Bus Management', hasAny('buses.read', 'buses.write', 'buses.verify', 'bus_owners.read', 'bus_owners.write', 'bus_owners.verify'));
    markGroup('Driver & Conductor Management', hasAny('drivers.read', 'drivers.write', 'drivers.verify', 'conductors.read', 'conductors.write', 'conductors.verify'));
    markGroup('Lounge Management', hasAny('lounges.read', 'lounges.write', 'lounges.verify', 'lounge_owners.read', 'lounge_owners.verify'));
    markGroup('Bus & Lounge Bookings', hasAny('bookings.read', 'bookings.write', 'lounge_bookings.read', 'lounge_bookings.write'));
    markGroup('Complaint', hasAny('complaints.read', 'complaints.update', 'complaints.escalate', 'escalation.read', 'escalation.manage'));
    markGroup('Users & Roles', hasAny('users.manage', 'permissions.manage'));
  }


  activeSessions = [
    { username: 'Dinesh Priyash', device: 'Chrome on Windows', workstation: '192.168.1.1', time: '2024-10-24 14:30', status: 'Active' },
    { username: 'Dinesh Priyash', device: 'Safari on iPhone', workstation: '192.168.1.5', time: '2024-10-23 09:15', status: 'Idle' },
    { username: 'Dinesh Priyash', device: 'Firefox on Mac', workstation: '192.168.1.8', time: '2024-10-22 18:45', status: 'Expired' }
  ];

  privacySettings = {
    profileVisibility: {
      adminsOnly: true,
      operationsManagers: false,
      bookingManagers: false,
      everyone: false
    },
    exportControls: {
      adminsOnly: true,
      operationsManagers: true,
      bookingManagers: false,
      everyone: false
    },
    privacyNotifications: {
      viewSensitiveInfo: true,
      downloadReport: true,
      updatePrivacy: true
    }
  };

  saveProfile() {
    console.log('Saving profile...', this.userProfile);
    // Implement save logic here
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

  saveSystemPreferences() {
    console.log('Saving system preferences:', this.systemPreferences);
    
    if (this.isBrowser) {
      const theme = this.systemPreferences.theme;
      this.originalTheme = theme; // Update original theme on save
      
      if (theme === 'system') {
        localStorage.removeItem('theme');
      } else {
        localStorage.setItem('theme', theme);
      }
    }
  }

  private applyTheme(theme: string) {
    if (theme === 'dark') {
      document.documentElement.setAttribute('data-theme', 'dark');
    } else {
      document.documentElement.removeAttribute('data-theme');
    }
  }

  private applySystemTheme() {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    this.applyTheme(prefersDark ? 'dark' : 'light');
  }

  cancelSystemPreferences() {
    console.log('Cancelling system preferences changes');
    // Revert to original theme
    this.selectTheme(this.originalTheme);
  }
}
