package store

// func TestSelectClips(t *testing.T) {
// 	ctx := context.Background()
// 	args := DefaultClipQueryArgs()
// 	clips, err := testDatabase.SelectClips(ctx, args)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	fmt.Printf("successfully retrieved %d clips from db\n", len(clips))
// }

// func TestSelectByGame(t *testing.T) {
// 	ctx := context.Background()
// 	args := DefaultClipQueryArgs()
// 	args.GameID = testGames[0].ID
// 	clips, err := testDatabase.SelectClips(ctx, args)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	for _, c := range clips {
// 		if c.GameID != args.GameID {
// 			t.Errorf("error: game ID mismatch")
// 		}
// 	}

// 	fmt.Printf("successfully retrieved %d clips from db by game id: %s\n", len(clips), args.GameID)
// }

// func TestSelectByBroadcaster(t *testing.T) {
// 	ctx := context.Background()
// 	args := DefaultClipQueryArgs()
// 	args.BroadcasterID = "37402112" // shroud
// 	clips, err := testDatabase.SelectClips(ctx, args)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	for _, c := range clips {
// 		if c.Broadcaster.ID != args.BroadcasterID {
// 			t.Errorf("error: broadcaster ID mismatch")
// 		}
// 	}

// 	fmt.Printf("successfully retrieved %d clips from db by broadcaster\n", len(clips))
// }

// func TestSelectByTimeRange(t *testing.T) {
// 	ctx := context.Background()
// 	args := DefaultClipQueryArgs()
// 	t1 := time.Now().Add(-time.Hour * 24 * 180)
// 	t2 := time.Now().Add(-time.Hour * 24 * 30)
// 	args.T1 = &t1
// 	args.T2 = &t2

// 	clips, err := testDatabase.SelectClips(ctx, args)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	for _, c := range clips {
// 		if c.ClipDate.Before(t1) {
// 			t.Errorf("clip_date entry: %s predates query parameter %s", c.ClipDate, t1)
// 		} else if c.ClipDate.After(t2) {
// 			t.Errorf("clip_date entry %s suceeds query paramter %s", c.ClipDate, t2)
// 		}
// 	}

// 	fmt.Printf("successfully retrieved %d clips by time range\n", len(clips))
// }

// func TestSelectWithOffset(t *testing.T) {
// 	ctx := context.Background()
// 	args := DefaultClipQueryArgs()
// 	args.Offset = 2

// 	clips, err := testDatabase.SelectClips(ctx, args)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	fmt.Printf("successfully retrieved %d clips by time range\n", len(clips))
// }
