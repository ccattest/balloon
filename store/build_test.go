package store

// func TestBuildVid(t *testing.T) {
// 	ctx := context.Background()
// 	args := DefaultClipQueryArgs()
// 	args.GameID = testGames[0].ID

// 	curDuration := time.Second * 0
// 	targetDuration := time.Minute * 5

// 	clips := []models.Clip{}

// 	for curDuration < targetDuration {
// 		res, err := testDatabase.SelectClips(ctx, args)
// 		if err != nil {
// 			t.Error(err)
// 		}

// 		for _, c := range res {
// 			curDuration = time.Duration(int(curDuration.Seconds())+c.Duration) * time.Second
// 			clips = append(clips, c)
// 			if curDuration > targetDuration {
// 				break
// 			}
// 		}
// 		args.Offset = len(res)
// 	}
// }
