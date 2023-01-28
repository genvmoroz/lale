package yourdictionary

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

//go:embed test-data/testbody.html.txt
var testBody string

func TestSentenceScraperScrapeSentences(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, err := fmt.Fprint(w, testBody)
				if err != nil {
					log.Panicln(err)
				}
			},
		),
	)
	defer srv.Close()

	type fields struct {
		host    string
		retries uint
		timeout time.Duration
	}
	type args struct {
		word string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "correct",
			fields: fields{
				host:    srv.URL,
				retries: 3,
				timeout: 10 * time.Second,
			},
			args: args{word: "hello"},
			want: []string{
				"Hello, can I help you with something?",
				"She then turned and sauntered towards Victor, not a hint of fear in her carriage \"Hello, Victor\" His greeting was returned with a hiss, \"Elisabeth!\"",
				"You can also use them when leaving comments for friends, resulting in memorable birthday greetings or just an everyday hello that's far from ordinary.",
				"Hello DanaFirst, let me say that I'm not a big makeup user.",
				"Young and the Restless comings and goings can mean saying hello and goodbye to fan favorites.",
				"The best way to use a hello cheer for introducing players is to have one cheerleader use a megaphone or loudspeaker to announce the players names and stats.",
				"Silence followed until Randy said, \"Hello?\"",
				"He shrugged as she ignored his hello smile.",
				"She wondered if she'd freak everyone out with a few Hello Kitty posters to take away the severity of the rooms.",
				"He cursed his luck while muttering a clipped, \"Hello.\"",
				"Darkyn didn't tear the fabric between worlds for a simple hello.",
				"Hello Michelle, I'm Jackson.",
				"Hello, I'm 13 and am in public middle school.",
				"When you are ready to order your invitations, be sure to check out Hello!Lucky and see what they have to offer.",
				"The past life test at Hello Quizzy is a little different.",
				"Female punks are opting for a girlish indulgence and Hello Kitty stickers, appliqués, and paraphernalia are exactly the accessories you need to add a touch of pink to your punk.",
				"The helpful addition of Hello Kitty pink stickers to your mannish boots provides a wonderful girlish spark to your otherwise androgynous attire.",
				"Say hello to your Ash Vavin High Tops, known for their careful styling, quality materials, and impressive shaping.",
				"Not to be tossed around like \"hello\" or \"goodbye,\" the term is a reminder of the spiritual root of yoga.",
				"Cheerleading hello cheers can be used in a variety of ways.",
				"With his first ever tweet he said, \"Hello World.",
				"Take a shower and come down to say hello to Gio.",
				"Hello, Jackson… Sarah… and what have we here, a human pet?",
				"He was quick about it and placed a Hello Kitty Band-Aid over the small puncture before dropping the vials into his coat pocket.",
				"Hello, Jacksonâ€¦ Sarahâ€¦ and what have we here, a human pet?",
				"Hello, I'm looking for some fresh asparagus Hello Mrs Depp.",
				"Say hello to mince pies from the in-store bakery.",
				"Regards, Lawrence Chard Q Hello there, My inquiry is I have a gold bangle.",
				"Bob immediately blurts, \" Hello, my name is Bob Hill, and this is my wife, Betty.",
				"What can Hello Kitty do for your punk shoes?",
				"Hello Arnie Toby has cuttlefish to nibble on and i sprinkle nutrobal regulary on his food.",
				"At home you'd be able to ignore it and say hello daddy long legs.",
				"Until next time ding DING Hello and welcome to another newsletter.",
				"Ahotz says, \" hi all \" Franco says, \" hello everybody \" Paolo finds its way in.",
				"Hello again forgot to say that my Son also has a 70's spacehopper made by the Wembley company.",
				"In Hannah's circle, the opening gambit is no longer \" Hello, how are you?",
				"Can one say \" hello \" to mean goodbye?",
				"Hello Everyone Hope everyone is enjoying the heat wave and aren't to sizzled!",
				"He would not say hello or even nod at me.",
				"So moving with all the gorgeous kids waving like crazy and shouting hello or goodbye and all wanting their photos taken.",
				"Hello again,We looked around to find more veterinarian hospitals in Kuwait, and we finally found one.",
				"Hello, My cat Phoebe, age four, has been systematically losing her rear end and hind quarter hair since August.",
				"The Hello, Goodbye Window by Norman Juster (the author of The Phantom Tollbooth) is the 2006 Caldecott Award winner.",
				"Say sawatdee krub (hello in Thai) to Thai food when you're in the mood for something different.",
				"With cartridges featuring Mickey Mouse, Hello Kitty, and Sesame Street as well as a number of sports and floral designs, the Cricut is a great resource for making handmade gifts.",
				"Smile at him, say hello, talk about the weather, school, whatever.",
				"Even though there seems to be lots of \"dos\" and \"don'ts\", you can figure out what is proper and what is inappropriate by reading the following interview with Sabrina and Eunice Moyle of Hello!Lucky Wedding Invitations.",
				"Cartoon Critters - This site is based on real cartoon character celebrities like Dora the Explorer, Hello Kitty and Scooby Doo.",
				"The Hello Kitty Tongue ring is similar to the belly ring, but meant to be worn in the mouth through a tongue piercing.",
				"Although Hello Kitty body jewelry is meant for the older crowd, Sanrio's focus is still mainly set on younger girls.",
				"The Hello Kitty charm bracelet pairs Kitty with cute enameled flowers and has plenty of links to hold charms.",
				"At the less expensive end of the spectrum, we have the Hello Kitty choker.",
				"Hello Kitty seems to hold lasting appeal for both girls and young women.",
				"Stores such as Wal-mart, Sam's Club and Meijer's hire store greeters to say hello to customers as they enter their store and guide the customer to any sales or special items.",
				"This might be more convenient for out-of-town guests or those with prior engagements who may still want to stop by for a quick hello.",
				"For those wanting a cute case, the Hello Kitty Contact Lens Case is the perfect option for you.",
				"It is done with a pink hinged dome case that features a strawberry design as well as Hello Kitty.",
				"The storage case has Hello Kitty on the left side and a R on the right.",
				"Say hello to Fran Drescher and Fred Durst when they get there.",
				"The standard system was a very light grey (some just call it white), with limited versions available in black, blue, Sonic Anniversary, and Hello Kitty designs.",
				"If you've just got to have that Hello Kitty-themed Dreamcast, this is the only place you'll ever find it.",
				"If you prefer cartoon characters, it usually isn't difficult to find a charm featuring Hello Kitty, Winnie the Pooh, or Pikachu.",
				"Knowing how to greet a Sri Lankan in their native language is important whether you're at a job interview, trying to sell a service or a product to a citizen of Sri Lanka, or just saying hello to your new neighbor.",
				"Hello, I am currently involved in an adjustable rate mortgage which is now adjusting to an alarming 15% interest rate.",
				"Hello, Citimorgage just increased my Escrow account payments by $500.",
				"Nancy Grace wore a memorable tee when she was expecting twins that said, \"Say hello to my lil' friend\".",
				"The site includes selections from numerous vendors like Roxy, Hello Kitty, Xhilaration, Puma and more.",
				"The preschool toy line also includes several fun designs, such as a Hello Kitty car for little girls.",
				"In one moment, as Willie is canoodling with Sue and she's begging him to stay in his Santa suit, The Kid walks in and calmly says, \"Hello Santa.",
				"Go with the second one, and say hello to that cute guy or girl you spot at the grocery store!",
				"Hello dear Lori, I have recently been dating a friend of about 6 months for the past few weeks.",
				"Say \"hello\" to people who you come in contact with.",
				"Hello Lori, I met and started dating this guy a week ago.",
				"Hello, I wish to remain anonymous, but I would like to tell you how much I love your advice!",
				"No kiss hello, no being by my side… he was mingling with every girl there besides me.",
				"Hello, I have been with a woman that is 3 years older than me for just past 4 months.",
				"For example, in some parts of the world the hand gesture for hello is offensive.",
				"Hello. I am currently \"in-between\" a relationship right now.",
				"Hello, I have been married about a year now and I am in the military.",
				"Kiss when saying \"goodnight\" and \"hello\".",
				"I would like it if sometimes we could hold hands or kiss each other hello and goodbye.\"",
				"If you introduce yourself by saying hello, the person may think you are just another lame person trying to hit on him/her.",
				"Notice if she smiles at you and says hello.",
				"In many cases, there's no \"line\" needed, but sometimes you may feel that you need something more exciting to say than a mere hello.",
				"A Chibimaru backpack highlights an adorable character from the beloved Hello Kitty cast.",
				"Touted as an active member of the Hello Kitty character gang, this little dog has his own red-roofed house and hankers after his favorite treats, a cookie-shaped, milk-flavored, ultra-yummy dog bone!",
				"Unfortunately for Hello Kitty character fans, the Chibimaru backpack is not currently sold by Sanrio; the company that manufactures Hello Kitty items.",
				"If you like Chibimaru, chances are you'll like some of the other Hello Kitty chracters as well.",
				"You can also shop by character and look for a wide assortment of other items from the Hello Kitty character cast.",
				"At Toys R Us, you'll find another Dora the Explorer 16\" Backpack, but this time the design is fashioned in the Hello Butterfly motif.",
				"Hello Beautiful Glamour Beach Bags - Hello Beautiful offers a photo gallery of some glamorous straw bags that feature leather handles and floral embellishments.",
				"You'll find clear, zippered totes in a number of primary colors, Hello Kitty plastic totes for the kids and several bags made from recycled plastic.",
				"Far from the Hello Kitty, Dora the Explorer, and Bob the Builder bedding you'll normally find for toddlers, Olive Kids offers colorful, vibrant patterns in classic themes.",
				"A great example of this is when Jerry Macguire says \"you had me at hello.\"",
				"Even if the shop is busy someone should be able to say \"hello\" and politely ask you to wait a minute.",
				"Whether you have a question about a tattoo design or you just want to contact your favorite tattoo artist from the show to say hello, you can use the contact function on the website to send a note directly to the studio.",
				"Aloha means \"hello,\" \"goodbye,\" and \"love.\"",
				"Instructor Wai Lana, whose yoga program is on many PBS stations in the United States, has a series of fitness yoga DVDs called Hello Fitness Yoga.",
				"Just as you need to learn to hold a pencil before you can write, autistic children need to be taught very basic social skills such as how to say hello, or even play with another child without any verbal interaction.",
				"Dream Kitty - If you are a Hello Kitty fan, this company offers over thirty different car accessories with that sweet kitty in mind.",
				"Have you ever considered making up a hello cheer for each cheerleader on your squad?",
				"You definitely don't want to perform a hello cheer every time the players play; that would take too long and would bore the crowd.",
				"However, making up \"hello cheers\" for the seniors on the team to introduce them for the big homecoming game or introducing positions with a hello cheer can really incite the crowd to cheer extra loud.",
				"Getting the crowd involved with some \"hello cheer tactics\" can definitely set the tone for the game.",
				"Challenge each class to come up with the loudest and most creative hello cheer to introduce their class and what they're all about.",
				"Remember that your cheerleading hello cheers set the tone for the entire game.",
				"Services such as Moo offer many variations on the printing idea including hello cheers, sticker books and gift cards.",
				"Try throwing it to the crowd during your hello cheer, or awarding candy to the loudest section in the gym.",
				"Hello Cheers are a kind of introduction of your squad to the fans, often introducing each member of the squad individually.",
				"Hello Knitty is a fun site with interesting patterns and a good collection of boutique yarns.",
				"For example, if you are scheduled to go on a business trip to France, you may want to begin with transportation phrases, or learn different ways to say hello.",
				"Some elements of the website require a subscription for access; however, most of the activities are available for free, and a subscription can be obtained free of charge by placing a link to Hello World on your own website.",
				"Hello Cleveland.com provides a directory of insurance brokers and agents in Ohio.",
				"This song is from their third album, Hello Love, and if you like this free track, there are three others where this one came from.",
				"These days, the term MP3 is thrown around as freely as Hello! in youthful circles, but those out there who haven't been as close to the cutting edge might still feel a bit out of the loop.",
				"For example, suppose a popular social networking website like MySpace or Facebook instituted a new practice of saying hello to another user with a virtual \"shoulder tap.\"",
				"They kissed warmly, not just a hello kiss.",
				"Complete the puzzles to reveal colorful Hello Kitty images!",
				"She was beaming as I waved hello to her and shouted \" They've got him.",
				"By the way I think jambo means hello or something similar.",
				"These strategies should also include a scheme for ' golden hello ' payments for new staff in shortage subjects.",
				"And everywhere we go, Michael gets a wave and a cheery hello from the people who are making things happen.",
				"A quick hello to my little sister, who told me how great the War Child album is.",
				"A big hello to anyone I was at school with.",
				"Staff will welcome you with a herbal drink and a friendly hello, and guide you through your treatments with expert grace and knowledge.",
				"I welcome you with open arms to say a warm hello, Do we have a future, will you stay or go?",
				"So, when compiling, I used \" g++ -I $ INCLUDE -o hello hello hello.c \" to compile.",
				"So, when compiling, I used \" g++ -I $ INCLUDE -o hello hello.c \" to compile.",
				"Cruise river uniworld and shout hello us to the.",
				"Daddy's boy As I enter the open-plan office at his Camden headquarters, Stelios waves hello.",
				"Thanks very much, webmaster@nmbva.co.uk Hello, I am looking to find any ex 4th hussars that served in Malaya from 1948 to 1951.",
				"The company achieved infamy three years ago when several of its advertisements were banned by Hello!",
				"The affectations of riot grrl, hello kitty, little dresses, cartoon girls replacing real heroes, started to grate.",
				"Hello all thinking of buying Samsung LE32R41BD from john lewis £ 995 with free 5 year guarantee is this a good deal?",
				"I'm here, I thought I'd... Hello, me old mucker!",
				"Hello Boys, Just to let you know I'm still around, in case I read my own obituary.",
				"Say hello to ' Snap ', our famous snapdragon, who was once part of the Norwich's historic pageantry.",
				"Joseph Hello, I'm Dave from Kent I have been a cliff Richard fan since as long as I can remember.",
				"This time can also include a regular activity like guessing a riddle or a song to say hello.",
				"Hello again, I have a ripper of sudoku for everyone to try.",
				"Meet and greet It's polite to open your message with a simple hello or use the person's name or other suitable salutation.",
				"Hello everyone, so far so good, We've got confirmation through from where we're having the whole shebang.",
				"Hello, is it too late to do a shook swarm?",
				"Bill would love to hear from anyone that remembers him Please contact Webmaster for further details Hello.",
				"Hello, I'm a strange weirdo, who's a bit of a contradiction.",
				"Joseph Hello, I 'm Dave from Kent I have been a cliff richard fan since as long as I can remember.",
				"The blue rinse brigade (hello Mavis) love a bit of Brucie on a Saturday night.",
				"Meet and greet It 's polite to open your message with a simple hello or use the person 's name or other suitable salutation.",
				"In conclusion can I say hello to all my friends in GA everywhere.",
				"Hello everyone, so far so good, We 've got confirmation through from where we 're having the whole shebang.",
				"Hello, We are professional manufacturer who can provide more than 150 AAAA grade Classical replicas tiffany jewelries at very favorable price.",
				"How can you print out the uppercase version of Hello, World !",
				"Best regards, Ruth Hello Thank you so much for the oil blue fitted v-neck sweater which arrived today.",
				"Bill would love to hear from anyone that remembers him Please contact webmaster for further details Hello.",
				"Hello and welcome to issue 2 of the weevil mag.",
				"Hello, I 'm a strange weirdo, who 's a bit of a contradiction.",
				"Hello John, The drivers for the wireless adapter may be available from the manufacturers website.",
				"Have an awesome year y'all, Linz x x x Welfare Officer - Justine Allan Hello all you clever clogs '.",
				"We met Gwyneth Paltrow's son Moses, welcomed Katie Holmes' daughter Suri, and said hello to Brooke Shields' daughter Grier.",
				"If a member comes around, smile, say hello, and then continue doing what you were doing.",
				"Using a graduation poem or prayer is a sentimental way to say goodbye to one part of life and hello to the future.",
				"Hello Mary,You're right -- if you don't have a current working or friendly relationship with this attorney, then the invitation was inappropriate for them to send.",
				"Fortunately, wedding invitation experts Eunice and Sabrina Moyle from Hello!Lucky Wedding Invitations have some great tips to help you.",
				"Be sure to visit their website, Hello!Lucky, for more information.",
				"Hello, Cupcake! by Karen Tack and Alan Richardson is one of the best cake decorating books on the market.",
				"He has mentioned in interviews that his song, You Had Me From Hello was inspired by Zellweger's famous line in Jerry Maguire.",
				"Hello! magazine, calls the woman's story \"amazing\" because, you know, it's amazing that she got pregnant with Jude Law's baby and he doesn't remember anything.",
				"Cats loathe getting wet, but Hello Kitty welcomes the chance to keep your girl sunny and bright on rain-filled days.",
				"Hello, I wonder if you could tell me whether it is okay to use a six-month-old Jack Russell at stud?",
				"Hello, I am writing this because I have a few issues.",
				"Hello,My ten-year-old Terrier/Spaniel is hyperventilating and she keeps coughing.",
				"Hello,I have an 18-month-old Lhasa Apso that was bred 33 days ago.",
				"I hope that answers your student's question, and please say hello to the class for me.",
				"Hello,It may be nothing but my Shih Tzu/Maltese puppy is acting strangely.",
				"Hello, We have a four year old Rottweiler that just had puppies in March of this year.",
				"Hello, I live in Northeast Florida and have a lovely vine which I planted four years ago that had a label of \"Japanese Jasmine\".",
				"Hello Kitty belly rings are the latest in a long line of products bearing this cultural icon's likeness.",
				"Hello Kitty quietly made her world debut nearly thirty years ago.",
				"She simply began appearing in novelty stores, and the rest is history.For the few who may not know who Hello Kitty is, she is an utterly adorable image of a little white cat.",
				"Hello Kitty also seems to enjoy fashion.",
				"The Hello Kitty product line is extensive, even if you rarely see it advertised.",
				"Hello Kitty belly rings epitomize this trend toward market focus on non-traditional age groups.",
				"Hello Kitty naval rings sold at MonsterSteel.com are available in a bar bell design.",
				"A curved bar or post made from 316L stainless steel sports a ball at both ends, one to hold the jewelry in place, and another that features Hello Kitty's image hanging directly in front of the naval.",
				"So, you can say goodbye to the old shoe and racing car but say hello to the segway and altoid tin.",
				"You can also find free printables of popular children's characters such as Alice in Wonderland, Spongebob Squarepants, Barbie, Hello Kitty and Scooby Doo.",
				"For girls other ideas would be Lizzie McGuire, Hello Kitty, Disney Princess or even the Powerpuff girls.",
				"Tell everyone I said hello and I'm being held hostage by one of the Others.",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Host:    tt.fields.host,
				Retries: tt.fields.retries,
				Timeout: tt.fields.timeout,
			}

			s, err := NewSentenceScraper(cfg)
			if err != nil {
				t.Fatalf("NewSentenceScraper() error = %v, wantErr %v", err, false)
			}

			got, err := s.ScrapeSentences(tt.args.word, uint32(len(tt.want)))

			if (err != nil) != tt.wantErr {
				t.Fatalf("ScrapeSentences() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("ScrapeSentences() got len = %v, want len %v", len(got), len(tt.want))
			}

			for i := 0; i < len(got); i++ {
				if got[i] != tt.want[i] {
					t.Errorf("ScrapeSentences() got = %v, want %v, index %d", got[i], tt.want[i], i)
				}
			}
		})
	}
}

// Hello, I am currently involved in an adjustable rate mortgage which is now adjusting to an alarming 15% interest rate.,
// Hello, I am currently involved in an adjustable rate mortgage which is now adjusting to an alarming 15%!i(MISSING)nterest rate.,
func BenchmarkSentenceScraperScrapeSentences(b *testing.B) {
	srv := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, err := fmt.Fprint(w, testBody)
				if err != nil {
					log.Panicln(err)
				}
			},
		),
	)
	defer srv.Close()

	cfg := Config{
		Host:    srv.URL,
		Retries: 3,
		Timeout: 10 * time.Second,
	}

	s, err := NewSentenceScraper(cfg)
	if err != nil {
		b.Fatalf("NewSentenceScraper() error = %v, wantErr %v", err, false)
	}

	for i := 0; i < b.N; i++ {
		_, _ = s.ScrapeSentences("hello", 1000)
	}
}
