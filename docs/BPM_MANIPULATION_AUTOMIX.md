# BPM Manipulation & Automixing Feature Exploration

## Overview

This document explores the implementation of real-time BPM manipulation and automixing capabilities for Sound Cistern.

## Part 1: BPM Manipulation (Time-Stretching/Pitch-Shifting)

### What It Is
BPM manipulation allows users to speed up or slow down tracks to match a target BPM without changing the pitch (time-stretching) or while changing the pitch (pitch-shifting).

### Technical Approaches

#### Option 1: Web Audio API + Custom Processing (Complex)
**Pros:**
- Full control over the algorithm
- No external dependencies
- Works offline

**Cons:**
- Extremely complex to implement well
- Quality varies significantly
- High CPU usage

**Implementation approach:**
```javascript
// Pseudo-code for time-stretching using Web Audio API
class TimeStretcher {
  constructor(audioContext) {
    this.audioContext = audioContext;
    this.source = null;
    this.targetBPM = 128;
    this.originalBPM = 120;
  }
  
  calculateSpeedRatio() {
    return this.targetBPM / this.originalBPM;
  }
  
  // Basic approach: just change playbackRate (changes pitch)
  applyPitchShift(audioBuffer) {
    const ratio = this.calculateSpeedRatio();
    const source = this.audioContext.createBufferSource();
    source.buffer = audioBuffer;
    source.playbackRate.value = ratio;
    return source;
  }
  
  // Advanced: Phase vocoder for time-stretching
  // This is where it gets very complex - requires:
  // 1. STFT (Short-Time Fourier Transform)
  // 2. Phase vocoder algorithm
  // 3. Overlap-add synthesis
}
```

#### Option 2: SoundTouch.js (Recommended)
**Library:** https://github.com/cutterkom/soundtouch-js or https://www.surina.net/soundtouch/

**Pros:**
- Industry-standard algorithms
- Good quality results
- Maintains pitch when time-stretching
- Well-tested and documented

**Cons:**
- Additional dependency (~100KB)
- Requires WASM or JS implementation
- Some artifacts at extreme ratios

**Implementation:**
```javascript
import { SoundTouch, SimpleFilter } from 'soundtouch-js';

class BPMController {
  constructor(audioContext) {
    this.audioContext = audioContext;
    this.soundTouch = new SoundTouch();
  }
  
  setTargetBPM(originalBPM, targetBPM) {
    const tempoRatio = targetBPM / originalBPM;
    this.soundTouch.tempo = tempoRatio;
    // pitch stays the same automatically
  }
  
  setPitchShift(semitones) {
    this.soundTouch.pitch = Math.pow(2, semitones / 12);
  }
  
  processAudio(audioBuffer) {
    // SoundTouch processes audio buffers
    // Returns modified buffer at new tempo
  }
}
```

#### Option 3: Web Audio API playbackRate (Simple)
**Pros:**
- Built into Web Audio API
- Zero additional dependencies
- Very low latency

**Cons:**
- Changes both tempo AND pitch together
- No option to preserve pitch

**Implementation:**
```javascript
// Already in your Player.jsx
audioElement.playbackRate = targetBPM / originalBPM;
```

### Recommended Implementation Strategy

For Sound Cistern, I recommend **Option 2 (SoundTouch.js)** for production quality, with a fallback to Option 3 for simplicity.

#### Architecture

```
Player.jsx
├── AudioContext
├── Source Node (BufferSource or MediaElement)
├── BPMController (SoundTouch wrapper)
│   ├── Input: Original BPM, Target BPM
│   ├── Processing: SoundTouch chain
│   └── Output: Modified audio stream
└── Master Gain -> Analyser -> Destination
```

#### UI Controls

Add to Player.jsx:
- BPM display with current vs original
- Slider to adjust target BPM (-20% to +20% range)
- "Sync BPM" button to match currently playing track
- "Reset" button to return to original

## Part 2: Automixing by BPM Matching

### What It Is
Automatically create playlists where tracks flow together by matching compatible BPMs and intelligently ordering them.

### Core Concepts

#### 1. Compatible BPM Ranges
Tracks are considered "mixable" if their BPMs are within a certain percentage:
- **Strict:** ±3% (professional DJ standard)
- **Normal:** ±6% (good for casual listening)
- **Loose:** ±12% (creative transitions)

#### 2. Transition Types

**Beatmatching:**
- Adjust outgoing track to match incoming track's BPM
- Crossfade during a breakdown or outro

**Energy Bridge:**
- Use tracks with compatible energy levels
- BPM gradually changes over multiple tracks

**Key Mixing (Advanced):**
- Consider musical key (Camelot wheel)
- Harmonically compatible transitions

### Implementation

#### Backend: Automix Algorithm

```go
// services/automix.go

package services

type AutomixConfig struct {
    TargetBPM       float64
    BPMTolerance    float64  // e.g., 0.06 for 6%
    MinTracks       int
    MaxTracks       int
    DurationMinutes int      // Target playlist length
}

type MixableTrack struct {
    Track       *models.Record
    BPM         float64
    Energy      float64  // derived from audio features
    Key         string   // musical key
    Transition  TransitionType
}

type TransitionType string

const (
    TransitionBeatmatch TransitionType = "beatmatch"
    TransitionFade      TransitionType = "fade"
    TransitionEnergy    TransitionType = "energy"
)

// GenerateAutomixPlaylist creates a playlist of compatible tracks
func GenerateAutomixPlaylist(
    app *pocketbase.PocketBase, 
    userID string, 
    config AutomixConfig,
) ([]MixableTrack, error) {
    
    // 1. Fetch all user's tracks with BPM data
    tracks, err := fetchTracksWithBPM(app, userID)
    if err != nil {
        return nil, err
    }
    
    // 2. Filter tracks within BPM tolerance of target
    compatibleTracks := filterByBPMRange(tracks, config.TargetBPM, config.BPMTolerance)
    
    // 3. Build compatibility graph
    // Nodes = tracks, Edges = compatibility score
    graph := buildCompatibilityGraph(compatibleTracks, config)
    
    // 4. Find optimal path through graph
    // Start with highest energy track at target BPM
    // Use greedy algorithm or simulated annealing
    playlist := generateOptimalSequence(graph, config)
    
    return playlist, nil
}

// Calculate how well two tracks mix together
func calculateMixCompatibility(t1, t2 *MixableTrack) float64 {
    score := 0.0
    
    // BPM similarity (closer = better)
    bpmDiff := math.Abs(t1.BPM - t2.BPM)
    bpmScore := 1.0 - (bpmDiff / math.Max(t1.BPM, t2.BPM))
    score += bpmScore * 0.4  // 40% weight
    
    // Energy match (similar energy = better flow)
    energyDiff := math.Abs(t1.Energy - t2.Energy)
    energyScore := 1.0 - energyDiff
    score += energyScore * 0.3  // 30% weight
    
    // Key compatibility (Camelot wheel)
    keyScore := calculateKeyCompatibility(t1.Key, t2.Key)
    score += keyScore * 0.3  // 30% weight
    
    return score
}
```

#### Frontend: Automix UI

```jsx
// components/AutomixGenerator.jsx

export function AutomixGenerator({ tracks }) {
  const [targetBPM, setTargetBPM] = useState(128);
  const [tolerance, setTolerance] = useState(6); // percent
  const [generating, setGenerating] = useState(false);
  const [playlist, setPlaylist] = useState(null);
  
  const generateMix = async () => {
    setGenerating(true);
    
    // Option 1: Client-side generation
    const mix = generateMixClientSide(tracks, { targetBPM, tolerance });
    
    // Option 2: Server-side generation
    // const mix = await fetch('/api/automix', {
    //   method: 'POST',
    //   body: JSON.stringify({ targetBPM, tolerance })
    // }).then(r => r.json());
    
    setPlaylist(mix);
    setGenerating(false);
  };
  
  return (
    <div className="automix-panel">
      <h3>AutoMix Generator</h3>
      
      <div className="controls">
        <label>Target BPM</label>
        <input 
          type="range" 
          min="80" max="180" 
          value={targetBPM}
          onChange={e => setTargetBPM(parseInt(e.target.value))}
        />
        <span>{targetBPM} BPM</span>
        
        <label>BPM Tolerance</label>
        <input
          type="range"
          min="3" max="12"
          value={tolerance}
          onChange={e => setTolerance(parseInt(e.target.value))}
        />
        <span>±{tolerance}%</span>
      </div>
      
      <button onClick={generateMix} disabled={generating}>
        {generating ? 'Generating...' : 'Generate Mix'}
      </button>
      
      {playlist && (
        <div className="mix-preview">
          <h4>Generated Playlist ({playlist.length} tracks)</h4>
          <div className="bpm-flow">
            {/* Visualize BPM flow */}
            {playlist.map((track, i) => (
              <div key={track.track_id} className="mix-node">
                <span className="bpm">{Math.round(track.bpm)} BPM</span>
                <span className="title">{track.track_title}</span>
                {i < playlist.length - 1 && (
                  <span className="transition">→</span>
                )}
              </div>
            ))}
          </div>
          <button onClick={() => saveAsPlaylist(playlist)}>
            Save as Playlist
          </button>
        </div>
      )}
    </div>
  );
}
```

### Client-Side Mix Generation Algorithm

```javascript
// utils/automix.js

export function generateAutomix(tracks, options) {
  const { targetBPM, tolerance, minTracks = 10 } = options;
  
  // 1. Filter tracks within BPM range
  const minBPM = targetBPM * (1 - tolerance / 100);
  const maxBPM = targetBPM * (1 + tolerance / 100);
  
  const compatibleTracks = tracks.filter(t => {
    const bpm = t.bpm || 0;
    return bpm >= minBPM && bpm <= maxBPM;
  });
  
  if (compatibleTracks.length < minTracks) {
    throw new Error(`Only ${compatibleTracks.length} tracks in BPM range`);
  }
  
  // 2. Sort by closeness to target BPM
  compatibleTracks.sort((a, b) => {
    const diffA = Math.abs((a.bpm || 0) - targetBPM);
    const diffB = Math.abs((b.bpm || 0) - targetBPM);
    return diffA - diffB;
  });
  
  // 3. Build playlist with smooth BPM transitions
  const playlist = [];
  const used = new Set();
  
  // Start with track closest to target BPM
  let currentBPM = targetBPM;
  
  for (let i = 0; i < Math.min(compatibleTracks.length, 20); i++) {
    // Find best next track (close BPM, not used, varied artists)
    const nextTrack = findBestNextTrack(
      compatibleTracks, 
      used, 
      currentBPM,
      playlist[playlist.length - 1]
    );
    
    if (!nextTrack) break;
    
    playlist.push(nextTrack);
    used.add(nextTrack.track_id);
    currentBPM = nextTrack.bpm || targetBPM;
  }
  
  return playlist;
}

function findBestNextTrack(candidates, used, currentBPM, previousTrack) {
  return candidates
    .filter(t => !used.has(t.track_id))
    .sort((a, b) => {
      let scoreA = 0, scoreB = 0;
      
      // BPM closeness
      scoreA -= Math.abs((a.bpm || 0) - currentBPM);
      scoreB -= Math.abs((b.bpm || 0) - currentBPM);
      
      // Avoid same artist twice in a row
      if (previousTrack && a.artist_name === previousTrack.artist_name) {
        scoreA -= 50;
      }
      if (previousTrack && b.artist_name === previousTrack.artist_name) {
        scoreB -= 50;
      }
      
      return scoreB - scoreA;
    })[0];
}
```

### Integration with Existing Features

#### 1. Enhanced Filter Panel
Add an "AutoMix Mode" toggle to the filter panel that:
- Enables BPM manipulation controls
- Shows compatible track highlighting
- Adds "Add to Mix" button on track cards

#### 2. Player Integration
When AutoMix mode is active:
- Automatically adjust BPM of next track to match current
- Crossfade at optimal transition points
- Display upcoming track with BPM preview

#### 3. Playlist System
Save generated mixes as playlists with metadata:
```json
{
  "id": "playlist_123",
  "name": "128 BPM House Mix",
  "type": "automix",
  "target_bpm": 128,
  "tolerance": 6,
  "tracks": [...],
  "transitions": [
    { "from": "track_1", "to": "track_2", "type": "beatmatch", "bpm": 128 }
  ]
}
```

### Implementation Roadmap

#### Phase 1: Basic BPM Manipulation (Week 1-2)
1. Integrate SoundTouch.js
2. Add BPM controls to player
3. Test with various tracks

#### Phase 2: BPM Filtering & Display (Week 2-3) ✅ Done
1. ✅ Extract BPM from API
2. ✅ Add BPM filters to UI
3. ✅ Show BPM on track cards

#### Phase 3: Simple AutoMix (Week 3-4)
1. Client-side playlist generation
2. BPM-compatible track suggestions
3. Save as playlist feature

#### Phase 4: Advanced AutoMix (Week 4-6)
1. Server-side algorithm
2. Key detection/matching
3. Transition point analysis
4. Automatic crossfading

### Libraries & Tools

**Audio Processing:**
- `soundtouch-js` - Time-stretching/pitch-shifting
- `wavesurfer.js` - Waveform visualization with BPM markers
- `essentia.js` - Audio analysis (BPM detection, key detection)

**Key Detection (Camelot Wheel):**
- `tonal` - Music theory library for key analysis
- Or use Essentia's key detection algorithms

**Waveform/BPM Visualization:**
- Custom canvas rendering
- Show beat grids
- Visual transition cues

### Performance Considerations

1. **Pre-computation:**
   - Analyze BPM/key on track import
   - Store in database
   - Cache analysis results

2. **Lazy Loading:**
   - Only load BPM manipulation when needed
   - Keep original audio file cached

3. **Web Workers:**
   - Process audio in background thread
   - Prevent UI blocking

### Conclusion

This feature set would position Sound Cistern as a unique platform combining:
- ✅ Music discovery (existing)
- ✅ Favorites management (existing) 
- ✅ BPM-based filtering (just implemented)
- 🎯 Real-time BPM manipulation (Phase 1)
- 🎯 Intelligent automixing (Phase 3-4)

The killer feature would be creating "infinite mixes" - automatically generating hours of continuous music that flows seamlessly based on user's taste and preferred BPM range.
