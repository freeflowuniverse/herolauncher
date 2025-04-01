use clap::{Parser, Subcommand};
use herojobs_client::{HeroJobsClient, Job, Result};

#[derive(Parser)]
#[command(name = "herojobs")]
#[command(about = "HeroJobs client", long_about = "Command-line client for HeroJobs Unix domain socket server")]
struct Cli {
    #[arg(short, long, default_value = "/tmp/herojobs.sock")]
    socket: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Submit a new job
    Submit {
        /// Circle ID
        #[arg(short, long, default_value = "testcircle")]
        circle: String,

        /// Topic
        #[arg(short, long, default_value = "default")]
        topic: String,

        /// Session key
        #[arg(short, long, default_value = "test-session")]
        session: String,

        /// HeroScript content
        #[arg(short = 'H', long)]
        heroscript: String,

        /// RhaiScript content
        #[arg(short, long, default_value = "")]
        rhaiscript: String,
    },

    /// Get a job by ID
    Get {
        /// Job ID
        #[arg(short, long)]
        jobid: String,
    },

    /// Delete a job by ID
    Delete {
        /// Job ID
        #[arg(short, long)]
        jobid: String,
    },

    /// List jobs
    List {
        /// Circle ID (optional)
        #[arg(short, long)]
        circle: Option<String>,

        /// Topic (optional)
        #[arg(short, long)]
        topic: Option<String>,
    },

    /// Get queue size
    QueueSize {
        /// Circle ID
        #[arg(short, long)]
        circle: String,

        /// Topic
        #[arg(short, long, default_value = "default")]
        topic: String,
    },

    /// Empty a queue
    QueueEmpty {
        /// Circle ID
        #[arg(short, long)]
        circle: String,

        /// Topic
        #[arg(short, long, default_value = "default")]
        topic: String,
    },

    /// Get a job from a queue without removing it
    QueueGet {
        /// Circle ID
        #[arg(short, long)]
        circle: String,

        /// Topic
        #[arg(short, long, default_value = "default")]
        topic: String,
    },

    /// Get and remove a job from a queue
    QueueFetch {
        /// Circle ID
        #[arg(short, long)]
        circle: String,

        /// Topic
        #[arg(short, long, default_value = "default")]
        topic: String,
    },
}

fn print_job(job: &Job) {
    println!("Job details:");
    println!("  Job ID: {}", job.jobid);
    println!("  Circle ID: {}", job.circleid);
    println!("  Topic: {}", job.topic);
    println!("  Status: {:?}", job.status);
    println!("  Time Scheduled: {}", job.time_scheduled);
    
    if job.time_start > 0 {
        println!("  Time Start: {}", job.time_start);
    }
    
    if job.time_end > 0 {
        println!("  Time End: {}", job.time_end);
    }
    
    if !job.error.is_empty() {
        println!("  Error: {}", job.error);
    }
    
    if !job.result.is_empty() {
        println!("  Result: {}", job.result);
    }
}

fn run() -> Result<()> {
    let cli = Cli::parse();
    let mut client = HeroJobsClient::new(&cli.socket);
    
    // Connect to server
    client.connect()?;
    
    // Execute command
    match cli.command {
        Commands::Submit { circle, topic, session, heroscript, rhaiscript } => {
            let job = client.create_job(&circle, &topic, &session, &heroscript, &rhaiscript)?;
            println!("Job submitted successfully:");
            print_job(&job);
        },
        
        Commands::Get { jobid } => {
            let job = client.get_job(&jobid)?;
            print_job(&job);
        },
        
        Commands::Delete { jobid } => {
            client.delete_job(&jobid)?;
            println!("Job deleted successfully");
        },
        
        Commands::List { circle, topic } => {
            let jobs = client.list_jobs(circle.as_deref(), topic.as_deref())?;
            println!("Jobs:");
            if jobs.is_empty() {
                println!("  No jobs found");
            } else {
                for job_id in jobs {
                    println!("  {}", job_id);
                }
            }
        },
        
        Commands::QueueSize { circle, topic } => {
            let size = client.queue_size(&circle, &topic)?;
            println!("Queue size: {}", size);
        },
        
        Commands::QueueEmpty { circle, topic } => {
            client.queue_empty(&circle, &topic)?;
            println!("Queue emptied successfully");
        },
        
        Commands::QueueGet { circle, topic } => {
            let job = client.queue_get(&circle, &topic)?;
            print_job(&job);
        },
        
        Commands::QueueFetch { circle, topic } => {
            let job = client.queue_fetch(&circle, &topic)?;
            print_job(&job);
        },
    }
    
    Ok(())
}

fn main() {
    if let Err(err) = run() {
        eprintln!("Error: {}", err);
        std::process::exit(1);
    }
}
