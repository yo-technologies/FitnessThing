#!/bin/bash

# Vertical Slices Architecture Demonstration
# This script shows the structure and benefits of the new architecture

echo "=== Vertical Slices Architecture Demonstration ==="
echo ""

echo "1. Feature-based directory structure:"
echo "   Each feature contains all its components in one place"
echo ""

if [ -d "internal/features" ]; then
    echo "Current vertical slices structure:"
    tree internal/features -I '__pycache__|*.pyc' 2>/dev/null || find internal/features -type d | sed 's/[^/]*\//  /g'
    echo ""
else
    echo "   [Features directory not found - run from backend directory]"
    echo ""
fi

echo "2. Shared components structure:"
echo "   Common utilities shared across all features"
echo ""

if [ -d "internal/shared" ]; then
    echo "Current shared components:"
    tree internal/shared -I '__pycache__|*.pyc' 2>/dev/null || find internal/shared -type d | sed 's/[^/]*\//  /g'
    echo ""
else
    echo "   [Shared directory not found - run from backend directory]"
    echo ""
fi

echo "3. Benefits of Vertical Slices:"
echo "   ✅ Better cohesion - all feature code in one place"
echo "   ✅ Reduced coupling - features don't share business logic layers"
echo "   ✅ Easier to modify - changes contained within feature boundary"
echo "   ✅ Parallel development - teams can work on different features"
echo "   ✅ Clear ownership - each feature has clear boundaries"
echo ""

echo "4. Example User Feature Slice:"
echo ""
if [ -d "internal/features/user" ]; then
    echo "User feature contains:"
    find internal/features/user -name "*.go" | head -10 | while read file; do
        echo "   📄 $file"
    done
    echo ""
else
    echo "   [User feature directory not found]"
    echo ""
fi

echo "5. Comparison with Traditional Layered Architecture:"
echo ""
echo "   Before (Layered):          After (Vertical Slices):"
echo "   internal/                  internal/"
echo "   ├── app/                  ├── features/"
echo "   ├── service/              │   ├── user/"
echo "   ├── repository/           │   ├── workout/"
echo "   └── domain/               │   └── exercise/"
echo "                             └── shared/"
echo ""

echo "6. Development Workflow with Vertical Slices:"
echo "   1. Choose a feature to work on (e.g., user management)"
echo "   2. Navigate to internal/features/user/"
echo "   3. All related code is in one place:"
echo "      - handlers/    (API layer)"
echo "      - service/     (business logic)"
echo "      - repository/  (data access)"
echo "   4. Make changes within the feature boundary"
echo "   5. Test the entire feature slice"
echo ""

echo "=== Architecture Migration Complete ==="
echo ""
echo "The backend has been restructured to use vertical slices architecture."
echo "Each feature is now self-contained and easier to develop and maintain."
echo ""
echo "For more details, see: VERTICAL_SLICES_ARCHITECTURE.md"