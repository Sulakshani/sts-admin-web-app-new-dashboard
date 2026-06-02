# TODO: Add Passenger Growth Column Chart to Dashboard

## Steps to Complete:

1. **Install Dependencies** ✅
   - Run `cd frontend/my-angular-app && npm install ng2-charts chart.js` to add ng2-charts and Chart.js for the column chart.

2. **Update Dashboard Component TypeScript** ✅
   - Import necessary modules: PassengerService, NgChartsModule, BaseChartDirective, Chart types.
   - Inject PassengerService into the constructor.
   - In ngOnInit, subscribe to passengers$ and compute monthly passenger counts for the current year.
   - Add chart properties: barChartData, barChartOptions, barChartType.

3. **Update Dashboard Component HTML** ✅
   - Add a new `.chart` div after the "Monthly Paid Revenue - Bus Booking" chart.
   - Include a header "Passenger Growth" and a canvas element with baseChart directive bound to the chart properties.

4. **Test the Implementation** ✅
   - Run `ng serve` in the frontend/my-angular-app directory.
   - Verify the chart renders on the dashboard with passenger data aggregated by months.
   - Check layout to ensure it's positioned on the right side of the revenue chart (may require minor SCSS adjustment if needed).

Progress: All steps completed.
