package types

// NewFeed creates a new feed instance
func NewFeed(
	signalID string,
	power int64,
	interval int64,
) Feed {
	return Feed{
		SignalID: signalID,
		Power:    power,
		Interval: interval,
	}
}

// NewFeedWithDeviation creates a new feed instance with deviation
func NewFeedWithDeviation(
	signalID string,
	power int64,
	interval int64,
	deviation int64,
) FeedWithDeviation {
	return FeedWithDeviation{
		SignalID:            signalID,
		Power:               power,
		Interval:            interval,
		DeviationBasisPoint: deviation,
	}
}

// NewCurrentFeeds creates a new current feeds instance
func NewCurrentFeeds(
	feeds []Feed,
	lastUpdateTimestamp int64,
	lastUpdateBlock int64,
) CurrentFeeds {
	return CurrentFeeds{
		Feeds:               feeds,
		LastUpdateTimestamp: lastUpdateTimestamp,
		LastUpdateBlock:     lastUpdateBlock,
	}
}

// NewCurrentFeedWithDeviations creates a new current feeds with deviations instance
func NewCurrentFeedWithDeviations(
	feedWithDeviations []FeedWithDeviation,
	lastUpdateTimestamp int64,
	lastUpdateBlock int64,
) CurrentFeedWithDeviations {
	return CurrentFeedWithDeviations{
		Feeds:               feedWithDeviations,
		LastUpdateTimestamp: lastUpdateTimestamp,
		LastUpdateBlock:     lastUpdateBlock,
	}
}

// CalculateInterval calculates feed interval from power
func CalculateInterval(power int64, powerStep int64, minInterval int64, maxInterval int64) (interval int64) {
	if power < powerStep {
		return 0
	}

	// divide power by power threshold to create steps
	powerFactor := power / powerStep

	interval = max(maxInterval/powerFactor, minInterval)

	return
}

// CalculateDeviation calculates feed deviation from power
func CalculateDeviation(power int64, powerStep int64, minDeviationBP int64, maxDeviationBP int64) (deviation int64) {
	if power < powerStep {
		return 0
	}

	// divide power by power threshold to create steps
	powerFactor := power / powerStep

	deviation = max(maxDeviationBP/powerFactor, minDeviationBP)

	return
}
