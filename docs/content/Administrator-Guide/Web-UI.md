---
sidebar_position: 4
tags: [administrator-guide, web-ui]
---

# Web UI Administration

COO-LLM includes a built-in web interface for monitoring and configuration management.

## Accessing the Web UI

The Web UI is available at the root path of your COO-LLM server:

```
http://localhost:2906/ui
```

### Authentication

Login with credentials from `server.webui` config:

These can be customized in your config:

```yaml
server:
  webui:
    admin_id: "myadmin"
    admin_password: "securepass123"
```

## Dashboard Overview

The main dashboard provides real-time insights into your COO-LLM deployment:

### Key Metrics

- **System Status**: Gateway operational status
- **Active Clients**: Number of clients with recent activity
- **Total Queries**: Number of API calls processed (last 7 days)
- **Total Cost**: Estimated API costs (last 7 days)

### Usage Analytics

View detailed usage analytics:

- **Daily Usage Chart**: Query volume over the last 7 days
- **Provider Distribution**: Pie chart showing query distribution by provider
- **Auto-refresh**: Configurable refresh intervals (10s, 30s, 1m, 5m, manual)

## Configuration Management

### Viewing Current Config

The Configuration page displays your current system configuration with syntax highlighting and organized sections.

### Configuration Sections

- **Server Settings**: Listen address, CORS, web UI credentials
- **Storage Configuration**: Runtime and config storage backends
- **Policy Settings**: Load balancing algorithm and caching

### Policy Configuration

Configure load balancing behavior:

- **Algorithm Selection**: Choose between round_robin, least_loaded, or hybrid
- **Priority Settings**: Set priority for provider selection (Latency, Cost, Availability, Quality)
- **Cache Settings**: Enable/disable response caching with TTL

## Monitoring & Metrics

### Metrics Dashboard

The Metrics page provides detailed analytics with interactive charts:

- **Summary Cards**: High-level overview of total requests, total tokens, total cost, and success rate.
- **Time Series Charts**: Latency, token usage, and cost over time
- **Provider Breakdown**: Metrics filtered by provider
- **Client Analysis**: Usage patterns by client
- **Date Range Selection**: Custom time periods with calendar picker
- **Filters**: Filter metrics by provider, provider key, and client. Also, change the display mode (Aggregated, By Provider, By Provider Key, By Client).
- **Data Table**: View detailed metrics data in a table.
- **Export Functionality**: Download data as CSV files.

### Statistics Overview

The Statistics page offers comprehensive system statistics with different views and groupings.

- **Summary Cards**: High-level overview of total requests, total tokens, total cost, and average latency.
- **Group By Filter**: Group statistics by Provider, Client, or Model.
- **Overview Tab**:
    - **Query Distribution Chart**: Distribution of queries by the selected group.
    - **Cost Distribution Chart**: Cost breakdown by the selected group.
- **Comparison Tab**:
    - **Performance Comparison Chart**: Compare latency and cost across different groups.
- **Raw Data Tab**:
    - **Raw Data View**: View the raw statistics data in JSON format.
    - **Detailed Table**: View detailed statistics in a table.

## Provider Monitoring

### Provider Analytics

Monitor provider performance through the Statistics page:

- **Usage Distribution**: Query volume by provider
- **Cost Analysis**: Spending breakdown by provider
- **Latency Tracking**: Response time monitoring
- **Error Rates**: Failure analysis per provider

### Key Performance

Track API key utilization:

- **Load Distribution**: Requests per key
- **Rate Limit Monitoring**: Usage vs. limits
- **Cost Attribution**: Spending per key
- **Health Status**: Key availability and errors

## Client Management

The Clients page provides comprehensive client analytics:

- **Summary Cards**: High-level overview of total clients, total queries, tokens consumed, and total cost.
- **Search and Filter**: Search for clients by ID or API key, and filter by status (active/inactive).
- **Sortable Table**: A detailed table of all clients with the following information:
    - Client ID
    - API Key
    - Total Queries
    - Tokens Consumed
    - Cost Incurred
    - Providers Used
    - Last Active
    - Status
- **Pagination**: The table is paginated to handle a large number of clients.
- **Export Data**: Download the filtered client data as a CSV file.

## System Administration

### Configuration Updates

Apply configuration changes through the web interface:

- **Live Updates**: Modify settings without server restart
- **Validation**: Automatic config validation before saving
- **Backup**: View current configuration state
- **Security**: Sensitive data masking in displays

### Data Export

Export system data for analysis:

- **Client Reports**: CSV export of client usage data
- **Metrics Data**: Historical metrics downloads
- **Configuration**: Export current config for backup

## Custom Web UI

COO-LLM supports serving custom web UI builds:

### Configuration

Specify custom UI path in config:

```yaml
server:
  webui:
    web_ui_path: "/path/to/custom/ui/build"
```

### Requirements

Custom UI builds should be static files (HTML, CSS, JS) that can be served by the embedded web server.

## Security Considerations

### Authentication

- **Strong Passwords**: Use complex admin passwords
- **Session Management**: Automatic session expiration
- **IP Restrictions**: Limit admin access by IP

### API Security

- **HTTPS**: Enable SSL/TLS in production
- **API Key Rotation**: Regularly rotate admin keys
- **Audit Logging**: Monitor admin actions

### Network Security

- **Firewall**: Restrict admin interface access
- **VPN**: Require VPN for admin access
- **Rate Limiting**: Protect against brute force attacks

## Troubleshooting Web UI

### Common Issues

**Login Fails**
- Check credentials in config
- Verify case sensitivity
- Clear browser cache

**Page Not Loading**
- Confirm server is running
- Check network connectivity
- Verify port configuration

**Configuration Not Saving**
- Validate YAML syntax
- Check file permissions
- Review server logs

**Metrics Not Updating**
- Check storage backend connectivity
- Verify Prometheus configuration
- Restart metrics collection

### Debug Mode

Enable debug logging for UI issues:

```yaml
logging:
  level: "debug"
```

Check browser developer console for client-side errors.