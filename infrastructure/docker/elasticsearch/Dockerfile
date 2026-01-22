FROM elasticsearch:8.16.1

# Note: Elasticsearch has built-in 'russian' analyzer (snowball stemmer)
# No additional plugins needed for basic Russian morphology

# Set default configuration
ENV discovery.type=single-node
ENV xpack.security.enabled=false
