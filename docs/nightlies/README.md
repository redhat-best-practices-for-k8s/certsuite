# GitHub Pages Workflow Dashboard

This creates a beautiful, automatically-updating dashboard that displays the
status of your GitHub Actions workflows. It's specifically designed to monitor
the QE testing workflows listed in `script/rekick-failed-workflows.sh`.

## 🎯 What It Shows

- **Success rates** for the last 3 days
- **Individual workflow status** for each OCP version (4.14-4.19)
- **Failed run counts** and quick access to debugging
- **Real-time updates** via scheduled GitHub Actions

## 🚀 Setup Instructions

### 1. Enable GitHub Pages

1. Go to your repository settings
2. Navigate to **Pages** in the left sidebar
3. Under **Source**, select "GitHub Actions"
4. Save the settings

### 2. Configure Permissions

The workflow needs these permissions (already configured in the workflow file):

- `contents: read` - To read repository files
- `actions: read` - To access workflow run data
- `pages: write` - To deploy to GitHub Pages
- `id-token: write` - For secure deployment

### 3. Deploy the Dashboard

#### Option A: Automatic Deployment

- The dashboard will automatically update every hour during business hours
  (9 AM - 6 PM UTC, Monday-Friday)
- It also updates whenever you push changes to the dashboard files

#### Option B: Manual Deployment

1. Go to the **Actions** tab in your repository
2. Select "Update Workflow Dashboard" workflow
3. Click "Run workflow" and choose the main branch

### 4. Access Your Dashboard

Once deployed, your dashboard will be available at:

```url
https://redhat-best-practices-for-k8s.github.io/certsuite/nightlies/
```

## 📊 Dashboard Features

### Summary Cards

- **Success Rate**: Overall pass/fail percentage
- **Failed Runs**: Number of failures in the last 3 days
- **Active Workflows**: Count of monitored workflows
- **Total Runs**: All workflow executions

### Workflow Details

Each workflow shows:

- Last 10 runs with status indicators (🟢 success, 🔴 failure, 🟡 pending)
- Direct links to failed runs for debugging
- Time since each run
- Success count and total runs

### Monitored Workflows

The dashboard tracks these workflows from your rekick script:

- QE OCP 4.14-4.19 Testing (regular & intrusive)
- qe-ocp-hosted.yml

## 🔧 Customization

### Change Update Frequency

Edit `.github/workflows/update-dashboard.yml`:

```yaml
schedule:
  # Run every 30 minutes instead of hourly
  - cron: '*/30 * * * *'
```

### Add More Workflows

Edit the `WORKFLOWS_TO_MONITOR` array in the workflow file:

```javascript
const WORKFLOWS_TO_MONITOR = [
  "QE OCP 4.14 Testing",
  "Your New Workflow Name",  // Add here
  // ... existing workflows
];
```

### Change Time Window

Modify `DAYS_BACK` to show more/fewer days:

```javascript
const DAYS_BACK = 7; // Show last 7 days instead of 3
```

## 🛠️ How It Works

1. **Scheduled GitHub Action** runs every hour during business hours
2. **Fetches workflow data** using the GitHub API
3. **Generates static HTML** with pre-populated data (avoids rate limits)
4. **Deploys to GitHub Pages** automatically
5. **Users see updated dashboard** without waiting for API calls

## 📱 Mobile Friendly

The dashboard is fully responsive and works great on:

- Desktop computers
- Tablets
- Mobile phones

## 🔍 Troubleshooting

### Dashboard Not Updating?

1. Check if GitHub Pages is enabled in repository settings
2. Verify the workflow has proper permissions
3. Look at the Actions tab for any workflow failures

### Missing Workflows?

The dashboard may not find workflows if:

- Workflow names don't match exactly
- Workflow files use different naming conventions
- Workflows are disabled

### Rate Limiting?

This shouldn't happen since we pre-generate the data, but if you see issues:

- The workflow uses the built-in `GITHUB_TOKEN`
- Rate limits are much higher for authenticated requests
- The update frequency can be reduced if needed

## 🎨 Styling

The dashboard uses a modern design with:

- Clean, minimal interface
- Color-coded status indicators
- Responsive grid layout
- Smooth animations and transitions

CSS custom properties make it easy to customize colors:

```css
:root {
  --success: #22c55e;  /* Green for successful runs */
  --failure: #ef4444;  /* Red for failed runs */
  --pending: #f59e0b;  /* Yellow for pending runs */
  --neutral: #6b7280;  /* Gray for neutral elements */
}
```
