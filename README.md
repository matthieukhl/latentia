# Latentia: AI-Powered SQL Optimization

An intelligent agent that automatically identifies slow SQL queries in your TiDB database and suggests optimized versions using AI. Built with Go, React, and powered by TiDB Serverless vector search.

## What It Does

Latentia monitors your database for slow-performing queries and uses artificial intelligence to suggest better, faster alternatives:

1. **Monitors** your TiDB database for slow queries automatically
2. **Analyzes** each query using AI and TiDB's documentation knowledge base
3. **Suggests** optimized SQL with clear explanations of improvements
4. **Verifies** suggestions by running performance tests
5. **Presents** results in a clean web interface for easy review

## Key Features

### Automatic Query Discovery

- Continuously monitors TiDB slow query logs
- No manual query submission required
- Focuses on queries that impact performance most

### AI-Powered Optimization

- Uses large language models to understand query patterns
- Leverages TiDB documentation for context-aware suggestions
- Applies database optimization best practices automatically

### Safe Verification

- Tests suggested optimizations before recommending them
- Runs performance comparisons using EXPLAIN ANALYZE
- Never modifies your actual data

### Clear Results

- Shows before/after performance metrics
- Explains what changed and why
- One-click approval for implementing suggestions

## How It Works

The system runs continuously in the background, so you can focus on building features while it handles database performance optimization.

## Quick Start

### Prerequisites

- TiDB database (TiDB Cloud or self-hosted)
- OpenAI API key (for AI analysis)
- Docker and Docker Compose

### Setup

1. **Clone and configure:**

   ```bash
   git clone git@github.com:matthieukhl/latentia.git
   cd latentia
   cp deploy/.env.sample .env
   ```

2. **Add your credentials to `.env`:**

   ```bash
   DB_DSN=your-tidb-connection-string
   OPENAI_API_KEY=your-openai-key
   ```

3. **Start the system:**

   ```bash
   docker compose up -d
   ```

4. **Open the web interface:**
   ```
   http://localhost:3000
   ```

That's it! The system will start monitoring your database and suggesting optimizations.

## Web Interface

The dashboard provides:

- **Slow Queries**: List of detected performance issues
- **Suggestions**: AI-generated optimizations with explanations
- **Performance**: Before/after metrics and improvement percentages
- **History**: Track of accepted optimizations and their impact

## Architecture

- **Backend**: Go with Gin web framework
- **Frontend**: React with TypeScript and Chakra UI
- **Database**: TiDB Serverless with vector search capabilities
- **AI**: OpenAI embeddings and language models
- **Deployment**: Docker containers with docker-compose

## Use Cases

### Development Teams

- Catch performance issues before they reach production
- Learn database optimization techniques from AI explanations
- Maintain query performance as your application grows

### Database Administrators

- Proactive monitoring without manual query analysis
- Evidence-based optimization recommendations
- Track performance improvements over time

### DevOps Teams

- Reduce database-related incidents
- Automate part of performance tuning workflow
- Get insights into application database usage patterns

## Safety & Security

- **Read-only analysis**: Never modifies your actual data
- **Sandboxed testing**: Optimizations tested in isolated environment
- **Manual approval**: All changes require human review
- **Audit trail**: Complete history of suggestions and decisions

## Contributing

Built for the Latentia Hackathon 2025. See `CLAUDE.md` for detailed technical documentation.

## License

MIT License - see [LICENSE](LICENSE) file for details.
