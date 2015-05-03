x = g.V("http://fr.wikipedia.org/wiki/Parti_lib%C3%A9ral_du_Qu%C3%A9bec").In("refers")

y = g.V("http://fr.wikipedia.org/wiki/Personnalit%C3%A9_politique").In("refers")

g.Emit(x.All().Intersect(y.All()))
